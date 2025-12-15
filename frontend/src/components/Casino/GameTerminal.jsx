import React, { useState, useEffect, useMemo } from 'react';
import idl from '../../idl/casino.json';
import { BN } from '@coral-xyz/anchor';
import { useConnection, useWallet } from '@solana/wallet-adapter-react';
import { WalletMultiButton } from '@solana/wallet-adapter-react-ui';
import { PublicKey, SystemProgram, Transaction } from '@solana/web3.js';
import * as anchor from "@coral-xyz/anchor";
import axios from 'axios';
import { Buffer } from 'buffer';
import HistoryPanel from "./HistoryPanel.jsx";

if (typeof window !== 'undefined') {
    window.Buffer = Buffer;
}

const API_URL = import.meta.env.PUBLIC_API_URL;

const PROGRAM_ID = new PublicKey("7WLsmcUxHVJ1hF6X1rVkLfRmaFWGi8Xdwqjys3mvqYxB");
const VAULT_ADDRESS = new PublicKey("Fak2ZEUZEsETSjNZJT5AkNYCipgfZxaSVjknnsVQ9QFX");

const SYMBOLS = ['😶', '🍒', '🍋', '🍇', '🍫', '🔔', '7️⃣', '💎', '🃏'];

export default function GameTerminal() {
    const { connection } = useConnection();
    const { publicKey, signTransaction, sendTransaction } = useWallet();

    const [logs, setLogs] = useState(["> System initialized...", "> Waiting for wallet connection..."]);
    const [balance, setBalance] = useState("0.00");
    const [grid, setGrid] = useState([[0,0,0],[0,0,0],[0,0,0]]);
    const [isSpinning, setIsSpinning] = useState(false);
    const [betAmount, setBetAmount] = useState(0.5);
    const [clientSeed, setClientSeed] = useState(Math.random().toString(36).substring(7));

    const addLog = (msg) => setLogs(prev => [`> ${msg}`, ...prev].slice(0, 10));

    const refreshBalance = async () => {
        if (!publicKey) return;
        try {
            const res = await axios.get(`${API_URL}/wallet/balance/${publicKey.toString()}`);
            setBalance(res.data.data.balance_sol);
        } catch (e) {
            console.error(e);
        }
    };

    const initSession = async () => {
        if (!publicKey) return;
        try {
            addLog("Initializing secure session...");
            await axios.post(`${API_URL}/game/session`, { wallet_address: publicKey.toString() });
            addLog("Session established.");
            refreshBalance();
        } catch (e) {
            addLog(`Error: ${e.response?.data?.error || e.message}`);
        }
    };

    const deposit = async () => {
        if (!publicKey) return;

        try {
            addLog("Initializing connection...");

            const idlObject = idl.default || idl;
            if (!idlObject.address) idlObject.address = PROGRAM_ID.toString();

            const provider = new anchor.AnchorProvider(connection, window.solana, {});
            const program = new anchor.Program(idlObject, provider);

            const [casinoVaultPDA] = PublicKey.findProgramAddressSync([Buffer.from("casino_vault")], PROGRAM_ID);
            const [userBalancePDA] = PublicKey.findProgramAddressSync([Buffer.from("user_balance"), publicKey.toBuffer()], PROGRAM_ID);

            const amountToDeposit = 1.0;
            addLog(`Preparing to deposit ${amountToDeposit} SOL...`);
            const lamports = new BN(amountToDeposit * 1_000_000_000);

            const instruction = await program.methods
                .deposit(lamports)
                .accounts({
                    casinoVault: casinoVaultPDA,
                    userBalance: userBalancePDA,
                    user: publicKey,
                    systemProgram: SystemProgram.programId,
                })
                .instruction();

            const transaction = new Transaction().add(instruction);

            const { blockhash, lastValidBlockHeight } = await connection.getLatestBlockhash();
            transaction.recentBlockhash = blockhash;
            transaction.feePayer = publicKey;

            const signedTx = await signTransaction(transaction);

            const signature = await connection.sendRawTransaction(signedTx.serialize(), { skipPreflight: true });

            addLog(`Tx Sent! Sig: ${signature.substring(0, 8)}...`);
            addLog("Waiting for confirmation (Do not close)...");

            try {
                await connection.confirmTransaction({
                    blockhash,
                    lastValidBlockHeight,
                    signature
                }, 'confirmed');
                addLog("✅ Blockchain confirmed.");
            } catch (confirmError) {
                console.warn("Confirmation timed out, but tx might have landed:", confirmError);
                addLog("⚠️ Confirmation timed out. Attempting sync anyway...");
            }

            addLog("Syncing with server...");

            try {
                await axios.post(`${API_URL}/wallet/sync`, {
                    wallet_address: publicKey.toString(),
                    tx_signature: signature
                });
                addLog("✅ Server Synced!");
                await refreshBalance();
            } catch (syncError) {
                addLog(`❌ Sync Error: ${syncError.response?.data?.error || syncError.message}`);
                addLog(`MANUAL RECOVERY: Send signature ${signature} to support.`);
            }

        } catch (e) {
            console.error(e);
            addLog(`❌ Transaction Aborted: ${e.message}`);
        }
    };

    const spin = async () => {
        if (!publicKey) return;
        setIsSpinning(true);
        addLog(`Spinning... Bet: ${betAmount} SOL`);

        try {
            const res = await axios.post(`${API_URL}/game/spin`, {
                wallet_address: publicKey.toString(),
                bet_amount: betAmount,
                client_seed: clientSeed
            });

            const data = res.data.data;

            let shuffle = 0;
            const interval = setInterval(() => {
                setGrid(g => g.map(r => r.map(() => Math.floor(Math.random() * 9))));
                shuffle++;
                if (shuffle > 10) {
                    clearInterval(interval);
                    setGrid(data.outcome);
                    setBalance(data.balance_after || balance);
                    addLog(data.is_win ? `WINNER! Payout: ${data.payout_sol} SOL` : "Result: Loss");
                    refreshBalance();
                    setIsSpinning(false);
                }
            }, 100);

        } catch (e) {
            setIsSpinning(false);
            addLog(`Spin Error: ${e.response?.data?.error || e.message}`);
        }
    };

    const withdraw = async () => {
        if (!publicKey) return;
        const previousBalance = balance;

        try {
            addLog("Requesting server authorization...");

            const res = await axios.post(`${API_URL}/wallet/withdraw`, {
                wallet_address: publicKey.toString(),
                amount: parseFloat(balance)
            });

            const { signature, recovery_id, nonce, amount_lamports } = res.data.data;

            setBalance("0.00");
            addLog("Auth received. Constructing transaction...");

            const sigBuffer = Buffer.from(signature, 'hex');
            const sigArray = Array.from(sigBuffer);

            const idlObject = idl.default || idl;
            if (!idlObject.address) idlObject.address = PROGRAM_ID.toString();

            const provider = new anchor.AnchorProvider(
                connection,
                window.solana,
                anchor.AnchorProvider.defaultOptions()
            );
            const program = new anchor.Program(idlObject, provider);

            const [casinoVaultPDA] = PublicKey.findProgramAddressSync([Buffer.from("casino_vault")], PROGRAM_ID);
            const [userBalancePDA] = PublicKey.findProgramAddressSync([Buffer.from("user_balance"), publicKey.toBuffer()], PROGRAM_ID);

            const instruction = await program.methods
                .withdraw(
                    new BN(amount_lamports),
                    new BN(nonce),
                    sigArray,
                    recovery_id
                )
                .accounts({
                    casinoVault: casinoVaultPDA,
                    userBalance: userBalancePDA,
                    user: publicKey,
                    systemProgram: SystemProgram.programId,
                })
                .instruction();

            const transaction = new Transaction().add(instruction);
            const { blockhash, lastValidBlockHeight } = await connection.getLatestBlockhash();
            transaction.recentBlockhash = blockhash;
            transaction.feePayer = publicKey;

            const signedTx = await signTransaction(transaction);

            const txSignature = await connection.sendRawTransaction(signedTx.serialize(), { skipPreflight: false });

            addLog(`Tx Sent! Sig: ${txSignature.substring(0, 8)}...`);
            addLog("Waiting for confirmation...");

            await connection.confirmTransaction({
                blockhash,
                lastValidBlockHeight,
                signature: txSignature
            }, 'confirmed');
            addLog("✅ Withdrawal Confirmed on Blockchain.");

            await axios.post(`${API_URL}/wallet/complete-withdraw`, {
                wallet_address: publicKey.toString()
            });

        } catch (e) {
            console.error("Withdraw Error:", e);

            addLog(`⚠️ Transaction failed/cancelled. Attempting refund...`);

            try {
                await axios.post(`${API_URL}/wallet/refund`, {
                    wallet_address: publicKey.toString()
                });
                addLog("✅ Refund Successful! Balance restored.");
            } catch (refundError) {
                console.error("Refund denied:", refundError);
                addLog("ℹ️ Refund denied. Checking if transaction actually succeeded...");

                try {
                    await axios.post(`${API_URL}/wallet/complete-withdraw`, {
                        wallet_address: publicKey.toString()
                    });
                    addLog("✅ It did succeed! State updated.");
                } catch (finalError) {
                    addLog("❌ Critical Sync Error. Contact Support.");
                }
            }

            await refreshBalance();
        }
    };

    useEffect(() => {
        if (publicKey) {
            addLog(`Connected: ${publicKey.toBase58().substring(0,6)}...`);
            initSession();
        }
    }, [publicKey]);

    return (
        <div className="terminal-interface">
            <div className="header-row">
                <WalletMultiButton />
                <div className="balance-display">
                    BAL: <span className="accent">{balance} SOL</span>
                </div>
            </div>

            <div className="slot-grid">
                {grid.map((row, rI) => (
                    <div key={rI} className="slot-row">
                        {row.map((cell, cI) => (
                            <div key={cI} className="slot-cell">
                                {SYMBOLS[cell] || '?'}
                            </div>
                        ))}
                    </div>
                ))}
            </div>

            <div className="controls">
                <div className="bet-control">
                    <span>BET: </span>
                    <input
                        type="number"
                        value={betAmount}
                        onChange={(e) => setBetAmount(parseFloat(e.target.value))}
                        step="0.1"
                        min="0.1"
                    />
                </div>
                <button onClick={spin} disabled={isSpinning || !publicKey}>
                    [ SPIN_REELS ]
                </button>
                <div className="secondary-actions">
                    <button onClick={deposit} disabled={!publicKey}>[ DEPOSIT ]</button>
                    <button onClick={withdraw} disabled={!publicKey}>[ WITHDRAW ]</button>
                </div>
            </div>

            <div className="logs">
                {logs.map((log, i) => (
                    <div key={i} className="log-entry">{log}</div>
                ))}
            </div>

            <HistoryPanel publicKey={publicKey} />

            <style>{`
                .terminal-interface {
                    border: 1px solid var(--border-color);
                    background: #000;
                    padding: 1rem;
                    font-family: 'Fira Code', monospace;
                    max-width: 600px;
                    margin: 0 auto;
                }
                .header-row {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 2rem;
                    border-bottom: 1px dashed var(--border-color);
                    padding-bottom: 1rem;
                }
                .accent { color: var(--accent-color); font-weight: bold; }
                
                .slot-grid {
                    display: flex;
                    flex-direction: column;
                    gap: 10px;
                    margin: 2rem 0;
                    background: #111;
                    padding: 20px;
                    border: 1px solid #333;
                }
                .slot-row {
                    display: flex;
                    justify-content: center;
                    gap: 10px;
                }
                .slot-cell {
                    width: 60px;
                    height: 60px;
                    border: 2px solid var(--accent-color);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    font-size: 2rem;
                    background: #000;
                    box-shadow: 0 0 10px rgba(0, 255, 157, 0.2);
                }

                .controls {
                    display: flex;
                    flex-direction: column;
                    gap: 1rem;
                    align-items: center;
                }
                button {
                    background: transparent;
                    border: 1px solid var(--accent-color);
                    color: var(--accent-color);
                    padding: 10px 20px;
                    font-family: inherit;
                    cursor: pointer;
                    font-weight: bold;
                    transition: all 0.2s;
                }
                button:hover:not(:disabled) {
                    background: var(--accent-color);
                    color: #000;
                }
                button:disabled {
                    border-color: #555;
                    color: #555;
                    cursor: not-allowed;
                }
                
                .logs {
                    margin-top: 2rem;
                    height: 150px;
                    overflow-y: auto;
                    font-size: 0.8rem;
                    color: #888;
                    border-top: 1px solid #333;
                    padding-top: 10px;
                }
                .log-entry { margin-bottom: 4px; }
            `}</style>
        </div>
    );
}