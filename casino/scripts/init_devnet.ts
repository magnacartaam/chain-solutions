import * as anchor from "@coral-xyz/anchor";
import { Program } from "@coral-xyz/anchor";
import { Casino } from "../target/types/casino";
import { PublicKey, Keypair, Connection, LAMPORTS_PER_SOL } from "@solana/web3.js";
import fs from "fs";

const AUTHORITY_HASH = [71,70,169,68,3,165,105,169,209,169,103,111,242,65,36,227,16,191,74,83,140,55,115,251,151,137,231,174,145,225,11,142];

async function main() {
    const connection = new Connection("https://api.devnet.solana.com", "confirmed");

    const serverJson = JSON.parse(fs.readFileSync("./server.json", "utf8"));
    const serverWallet = Keypair.fromSecretKey(new Uint8Array(serverJson));

    const wallet = new anchor.Wallet(serverWallet);
    const provider = new anchor.AnchorProvider(connection, wallet, {
        preflightCommitment: "confirmed",
    });
    anchor.setProvider(provider);

    const program = anchor.workspace.Casino as Program<Casino>;

    console.log("--- DEBUG INFO ---");
    console.log("Connected to:", connection.rpcEndpoint);

    const balance = await connection.getBalance(serverWallet.publicKey);
    console.log(`Server Wallet (${serverWallet.publicKey.toString()}) Balance:`, balance / LAMPORTS_PER_SOL, "SOL");

    if (balance < 0.1 * LAMPORTS_PER_SOL) {
        throw new Error("Server wallet has insufficient funds! Please airdrop to: " + serverWallet.publicKey.toString());
    }

    const [casinoVaultPDA] = PublicKey.findProgramAddressSync(
        [Buffer.from("casino_vault")],
        program.programId
    );

    console.log("------------------");
    console.log("Initializing Casino Vault...");
    console.log("Authority Hash:", Buffer.from(AUTHORITY_HASH).toString('hex'));

    try {
        const tx = await program.methods
            .initialize(AUTHORITY_HASH)
            .accounts({
                payer: serverWallet.publicKey,
            })
            .signers([serverWallet])
            .rpc();

        console.log("✅ Initialization complete!");
        console.log("Transaction Signature:", tx);
        console.log("Vault Address:", casinoVaultPDA.toString());
    } catch (err) {
        console.error("❌ Transaction Failed:", err);
        if (err.logs) {
            console.log("Logs:", err.logs);
        }
    }
}

main().catch(console.error);