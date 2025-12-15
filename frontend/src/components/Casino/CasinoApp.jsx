import React, { useMemo } from 'react';
import { ConnectionProvider, WalletProvider } from '@solana/wallet-adapter-react';
import { WalletAdapterNetwork } from '@solana/wallet-adapter-base';
import { PhantomWalletAdapter } from '@solana/wallet-adapter-wallets';
import { WalletModalProvider } from '@solana/wallet-adapter-react-ui';
import { clusterApiUrl } from '@solana/web3.js';

import GameTerminal from './GameTerminal';

import '@solana/wallet-adapter-react-ui/styles.css';

export default function CasinoApp() {
    const network = WalletAdapterNetwork.Devnet;
    const endpoint = useMemo(() => {
        if (import.meta.env.DEV) {
            return "http://localhost/solana-rpc";
        }
        return clusterApiUrl(network);
    }, [network]);
    const wallets = useMemo(() => [new PhantomWalletAdapter()], [network]);

    return (
        <ConnectionProvider endpoint={endpoint}>
            <WalletProvider wallets={wallets} autoConnect>
                <WalletModalProvider>
                    <GameTerminal />
                </WalletModalProvider>
            </WalletProvider>
        </ConnectionProvider>
    );
}