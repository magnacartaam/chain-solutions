import * as anchor from "@coral-xyz/anchor";
import { Program } from "@coral-xyz/anchor";
import { Casino } from "../target/types/casino";
import {PublicKey, Keypair, Connection} from "@solana/web3.js";
import fs from "fs";

const NEW_AUTHORITY_HASH = [71,70,169,68,3,165,105,169,209,169,103,111,242,65,36,227,16,191,74,83,140,55,115,251,151,137,231,174,145,225,11,142];

async function main() {
    const connection = new Connection("https://api.devnet.solana.com", "confirmed");

    const serverJson = JSON.parse(fs.readFileSync("./server.json", "utf8"));
    const serverWallet = Keypair.fromSecretKey(new Uint8Array(serverJson));

    const wallet = new anchor.Wallet(serverWallet);
    const provider = new anchor.AnchorProvider(connection, wallet, {
        preflightCommitment: "confirmed"
    });
    anchor.setProvider(provider);

    const program = anchor.workspace.Casino as Program<Casino>;


    const [casinoVaultPDA] = PublicKey.findProgramAddressSync(
        [Buffer.from("casino_vault")],
        program.programId
    );

    console.log("Updating Signing Authority on Vault:", casinoVaultPDA.toString());

    const tx = await program.methods
        .updateSigningAuthority(NEW_AUTHORITY_HASH)
        .accounts({
            authority: serverWallet.publicKey,
        })
        .signers([serverWallet])
        .rpc();

    console.log("✅ Authority Updated! Tx:", tx);
}

main().catch(console.error);