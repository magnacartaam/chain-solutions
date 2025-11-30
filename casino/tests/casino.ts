import * as anchor from "@coral-xyz/anchor";
import { Program } from "@coral-xyz/anchor";
import { Casino } from "../target/types/casino";
import { assert } from "chai";
import {
    Keypair,
    SystemProgram,
    PublicKey,
    LAMPORTS_PER_SOL,
} from "@solana/web3.js";
import { randomBytes } from "crypto";
import { keccak_256 } from "js-sha3";
import * as secp256k1 from "secp256k1";

describe("casino", () => {
    const provider = anchor.AnchorProvider.env();
    anchor.setProvider(provider);
    const program = anchor.workspace.Casino as Program<Casino>;

    const payer = provider.wallet as anchor.Wallet;

    let serverSecpPrivateKey;
    do {
        serverSecpPrivateKey = randomBytes(32);
    } while (!secp256k1.privateKeyVerify(serverSecpPrivateKey));
    const serverSecpPublicKey = Buffer.from(secp256k1.publicKeyCreate(serverSecpPrivateKey, false));

    const authorityHash = Buffer.from(keccak_256.digest(serverSecpPublicKey.slice(1)));

    const player = Keypair.generate();
    const [casinoVaultPDA] = PublicKey.findProgramAddressSync([Buffer.from("casino_vault")], program.programId);
    const [playerBalancePDA] = PublicKey.findProgramAddressSync([Buffer.from("user_balance"), player.publicKey.toBuffer()], program.programId);

    before(async () => {
        await provider.connection.requestAirdrop(player.publicKey, 10 * LAMPORTS_PER_SOL).then(sig => provider.connection.confirmTransaction(sig));
        await provider.connection.requestAirdrop(payer.publicKey, 10 * LAMPORTS_PER_SOL).then(sig => provider.connection.confirmTransaction(sig));
    });

    it("Initializes the casino vault!", async () => {
        await program.methods
            .initialize(Array.from(authorityHash))
            .accounts({
                casinoVault: casinoVaultPDA,
                payer: payer.publicKey,
                systemProgram: SystemProgram.programId,
            })
            .rpc();
        const fetchedVault = await program.account.casinoVault.fetch(casinoVaultPDA);
        assert.deepStrictEqual(Buffer.from(fetchedVault.signingAuthority), authorityHash);
        console.log("✅ Casino Initialized with correct signing authority HASH.");
    });

    it("Player makes a deposit", async () => {
        const depositAmount = new anchor.BN(LAMPORTS_PER_SOL);
        await program.methods.deposit(depositAmount).accounts({
            casinoVault: casinoVaultPDA,
            userBalance: playerBalancePDA,
            user: player.publicKey,
            systemProgram: SystemProgram.programId,
        }).signers([player]).rpc();
        const fetchedBalance = await program.account.userBalance.fetch(playerBalancePDA);
        assert.ok(fetchedBalance.amount.eq(depositAmount));
        console.log("✅ Player deposited 1 SOL.");
    });

    it("Player makes a valid withdrawal", async () => {
        const withdrawAmount = new anchor.BN(0.5 * LAMPORTS_PER_SOL);
        const nonce = new anchor.BN(1);

        const message = Buffer.concat([
            player.publicKey.toBuffer(),
            withdrawAmount.toBuffer('le', 8),
            nonce.toBuffer('le', 8),
        ]);
        const messageHash = Buffer.from(keccak_256.digest(message));

        const { signature, recid: recoveryId } = secp256k1.ecdsaSign(messageHash, serverSecpPrivateKey);

        await program.methods
            .withdraw(withdrawAmount, nonce, Array.from(signature), recoveryId)
            .accounts({
                casinoVault: casinoVaultPDA,
                userBalance: playerBalancePDA,
                user: player.publicKey,
            })
            .signers([player])
            .rpc();

        const fetchedBalance = await program.account.userBalance.fetch(playerBalancePDA);
        const expectedBalance = new anchor.BN(0.5 * LAMPORTS_PER_SOL);
        assert.ok(fetchedBalance.amount.eq(expectedBalance));
        assert.ok(fetchedBalance.lastWithdrawalNonce.eq(nonce));
        console.log("✅ Player withdrew 0.5 SOL.");
    });

    it("Server commits a batch root", async () => {
        const batchId = new anchor.BN(1);
        const merkleRoot = randomBytes(32);
        const [batchCommitPDA] = PublicKey.findProgramAddressSync([Buffer.from("batch_commit"), batchId.toBuffer('le', 8)], program.programId);

        await program.methods.commitBatchRoot(batchId, Array.from(merkleRoot)).accounts({
            casinoVault: casinoVaultPDA,
            batchCommit: batchCommitPDA,
            authority: payer.publicKey,
            systemProgram: SystemProgram.programId,
        })
            .rpc();

        const fetchedBatch = await program.account.batchCommit.fetch(batchCommitPDA);
        assert.ok(Buffer.from(fetchedBatch.merkleRoot).equals(merkleRoot));
        console.log("✅ Server successfully committed a Merkle root for Batch #1.");
    });
});