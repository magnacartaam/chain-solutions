import * as anchor from "@coral-xyz/anchor";
import { Program } from "@coral-xyz/anchor";
import { Casino } from "../target/types/casino";
import { PublicKey, Keypair, Connection, LAMPORTS_PER_SOL } from "@solana/web3.js";
import fs from "fs";

const AUTHORITY_HASH = [60,162,232,94,124,58,115,29,36,26,80,238,73,104,166,114,201,188,223,83,236,253,51,93,234,65,51,232,41,120,135,114,139,24,85,105,84,71,239,183,160,190,15,75,23,116,213,159,215,101,155,235,3,189,150,193,234,32,87,61,109,27,152,129];

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