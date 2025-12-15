import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Buffer } from 'buffer';
import { sha256 } from 'js-sha256';

const API_URL = import.meta.env.PUBLIC_API_URL;

export default function HistoryPanel({ publicKey }) {
    const [spins, setSpins] = useState([]);
    const [selectedSpin, setSelectedSpin] = useState(null);
    const [verificationStatus, setVerificationStatus] = useState(null);

    useEffect(() => {
        if (!publicKey) return;
        axios.get(`${API_URL}/game/history?wallet_address=${publicKey.toString()}`)
            .then(res => setSpins(res.data.data || []))
            .catch(err => console.error("History fetch error:", err));
    }, [publicKey]);

    const verifySpin = async (spinId) => {
        setVerificationStatus("Fetching proof...");
        try {
            const res = await axios.get(`${API_URL}/game/proof/${spinId}`);
            const data = res.data.data;

            const calcHash = sha256(data.spin.server_seed);
            const isSeedValid = calcHash === data.spin.server_seed_hash;

            let isMerkleValid = false;
            let calculatedRoot = "";

            if (data.merkle_root && data.proof) {
                const outcomeStr = JSON.stringify(data.spin.outcome.reels);

                const canonicalString = `${data.spin.wallet_address}:${data.spin.spin_nonce}:${data.spin.server_seed}:${data.spin.client_seed}:${data.spin.bet_amount}:${outcomeStr}:${data.spin.payout_amount}`;

                let currentHash = sha256(canonicalString);

                if (currentHash !== data.spin.leaf_hash) {
                    console.error("Local calculation differs from DB leaf hash. Canonical string mismatch?");
                    console.log("Local String:", canonicalString);
                }

                isMerkleValid = (currentHash === data.spin.leaf_hash);
            }

            setSelectedSpin({
                ...data,
                isSeedValid,
                isMerkleValid,
                localLeafHash: calculatedRoot // Debug info
            });
            setVerificationStatus("Done");

        } catch (e) {
            console.error(e);
            setVerificationStatus(`Error: ${e.response?.data?.error || e.message}`);
        }
    };

    return (
        <div className="history-panel">
            <h3>Recent Transactions</h3>
            <div className="table-container">
                <table>
                    <thead>
                    <tr>
                        <th>Time</th>
                        <th>Bet</th>
                        <th>Outcome</th>
                        <th>Payout</th>
                        <th>Action</th>
                    </tr>
                    </thead>
                    <tbody>
                    {spins.map(s => (
                        <tr key={s.spin_id} className={parseFloat(s.payout_amount) > 0 ? "win" : "loss"}>
                            <td>{new Date(s.created_at).toLocaleTimeString()}</td>
                            <td>{s.bet_amount} SOL</td>
                            <td>{parseFloat(s.payout_amount) > 0 ? "WIN" : "LOSS"}</td>
                            <td>{s.payout_amount}</td>
                            <td>
                                <button onClick={() => verifySpin(s.spin_id)}>[ VERIFY ]</button>
                            </td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>

            {/* Verification Modal / Details Panel */}
            {selectedSpin && (
                <div className="verify-details">
                    <h4>Verification Results</h4>
                    <div className="check-row">
                        <span>Server Seed Hash:</span>
                        <span className={selectedSpin.isSeedValid ? "green" : "red"}>
                            {selectedSpin.isSeedValid ? "✅ MATCHED" : "❌ FAILED"}
                        </span>
                    </div>
                    <div className="check-row">
                        <span>Blockchain Anchor:</span>
                        {selectedSpin.solana_tx ? (
                            <a
                                href={`https://explorer.solana.com/tx/${selectedSpin.solana_tx}?cluster=devnet`}
                                target="_blank"
                                rel="noreferrer"
                                className="green"
                            >
                                ✅ VIEW ON CHAIN (Batch #{selectedSpin.batch_id})
                            </a>
                        ) : (
                            <span className="yellow">⏳ PENDING (Wait for batch)</span>
                        )}
                    </div>
                    <button onClick={() => setSelectedSpin(null)}>[ CLOSE ]</button>
                </div>
            )}

            <style>{`
                .history-panel { margin-top: 2rem; border-top: 1px dashed #333; padding-top: 1rem; }
                table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
                th { text-align: left; color: #666; padding: 5px; }
                td { padding: 5px; border-bottom: 1px solid #222; }
                .win { color: var(--accent-color); }
                .loss { color: #555; }
                button { background: #111; border: 1px solid #444; color: #888; cursor: pointer; }
                button:hover { color: #fff; border-color: #fff; }
                
                .verify-details {
                    background: #111; border: 1px solid var(--accent-color);
                    padding: 1rem; margin-top: 1rem;
                }
                .check-row { display: flex; justify-content: space-between; margin-bottom: 0.5rem; }
                .green { color: var(--accent-color); }
                .red { color: #ff5f56; }
                .yellow { color: #ffbd2e; }
            `}</style>
        </div>
    );
}