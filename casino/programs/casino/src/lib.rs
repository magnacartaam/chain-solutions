use anchor_lang::prelude::*;
use solana_program::keccak;
use solana_program::secp256k1_recover::{secp256k1_recover};

declare_id!("7WLsmcUxHVJ1hF6X1rVkLfRmaFWGi8Xdwqjys3mvqYxB");

#[program]
pub mod casino {
    use super::*;

    pub fn initialize(ctx: Context<Initialize>, signing_authority: [u8; 32]) -> Result<()> {
        ctx.accounts.casino_vault.operational_authority = ctx.accounts.payer.key();
        ctx.accounts.casino_vault.signing_authority = signing_authority;
        ctx.accounts.casino_vault.batch_id_counter = 0;
        Ok(())
    }

    pub fn deposit(ctx: Context<Deposit>, amount: u64) -> Result<()> {
        let transfer_instruction = anchor_lang::system_program::Transfer {
            from: ctx.accounts.user.to_account_info(),
            to: ctx.accounts.casino_vault.to_account_info(),
        };
        let cpi_context = CpiContext::new(
            ctx.accounts.system_program.to_account_info(),
            transfer_instruction,
        );
        anchor_lang::system_program::transfer(cpi_context, amount)?;

        let user_balance = &mut ctx.accounts.user_balance;

        if user_balance.user == Pubkey::default() {
            msg!("New user account detected. Initializing.");
            user_balance.user = ctx.accounts.user.key();
            user_balance.last_withdrawal_nonce = 0;
            user_balance.bump = ctx.bumps.user_balance;
        }

        user_balance.amount = user_balance.amount.checked_add(amount)
            .ok_or(MyError::DepositOverflow)?;

        Ok(())
    }

    pub fn withdraw(ctx: Context<Withdraw>, amount: u64, nonce: u64, signature: [u8; 64], recovery_id: u8) -> Result<()> {
        require!(nonce == ctx.accounts.user_balance.last_withdrawal_nonce + 1, MyError::InvalidNonce);

        let mut message = Vec::new();
        message.extend_from_slice(&ctx.accounts.user.key().to_bytes());
        message.extend_from_slice(&amount.to_le_bytes());
        message.extend_from_slice(&nonce.to_le_bytes());
        let message_hash = keccak::hash(&message).to_bytes();

        let recovered_pubkey_bytes = secp256k1_recover(&message_hash, recovery_id, &signature)
            .map_err(|_| MyError::SignatureVerificationFailed)?.to_bytes();

        let recovered_authority_hash = keccak::hash(&recovered_pubkey_bytes).to_bytes();

        require!(recovered_authority_hash == ctx.accounts.casino_vault.signing_authority, MyError::SignatureVerificationFailed);

        require!(ctx.accounts.user_balance.amount >= amount, MyError::InsufficientUserBalance);
        let user_balance = &mut ctx.accounts.user_balance;
        user_balance.last_withdrawal_nonce = nonce;
        user_balance.amount -= amount;
        let vault_balance = ctx.accounts.casino_vault.to_account_info().lamports();
        require!(vault_balance >= amount, MyError::InsufficientVaultBalance);
        **ctx.accounts.casino_vault.to_account_info().try_borrow_mut_lamports()? -= amount;
        **ctx.accounts.user.to_account_info().try_borrow_mut_lamports()? += amount;

        Ok(())
    }

    pub fn commit_batch_root(ctx: Context<CommitBatch>, batch_id: u64, merkle_root: [u8; 32]) -> Result<()> {
        require_keys_eq!(ctx.accounts.authority.key(), ctx.accounts.casino_vault.operational_authority, MyError::Unauthorized);

        ctx.accounts.casino_vault.batch_id_counter += 1;

        let batch_commit = &mut ctx.accounts.batch_commit;
        batch_commit.authority = ctx.accounts.authority.key();
        batch_commit.batch_id = batch_id;
        batch_commit.merkle_root = merkle_root;
        Ok(())
    }
}

#[account]
pub struct CasinoVault {
    pub operational_authority: Pubkey,
    pub signing_authority: [u8; 32],
    pub batch_id_counter: u64,
}

#[account]
pub struct UserBalance {
    pub user: Pubkey,
    pub amount: u64,
    pub last_withdrawal_nonce: u64,
    pub bump: u8,
}

#[account]
pub struct BatchCommit {
    pub authority: Pubkey,
    pub batch_id: u64,
    pub merkle_root: [u8; 32],
}

#[derive(Accounts)]
pub struct Initialize<'info> {
    #[account(
        init,
        payer = payer,
        space = 8 + 32 + 32 + 8,
        seeds = [b"casino_vault"],
        bump
    )]
    pub casino_vault: Account<'info, CasinoVault>,

    #[account(mut)]
    pub payer: Signer<'info>,

    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct Deposit<'info> {
    #[account(mut)]
    pub casino_vault: Account<'info, CasinoVault>,

    #[account(
        init_if_needed,
        payer = user,
        space = 8 + 32 + 8 + 8 + 1,
        seeds = [b"user_balance", user.key().as_ref()],
        bump
    )]
    pub user_balance: Account<'info, UserBalance>,

    #[account(mut)]
    pub user: Signer<'info>,

    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct Withdraw<'info> {
    #[account(
        mut,
        seeds = [b"casino_vault"],
        bump
    )]
    pub casino_vault: Account<'info, CasinoVault>,

    #[account(
        mut,
        seeds = [b"user_balance", user.key().as_ref()],
        bump = user_balance.bump,
        constraint = user_balance.user == user.key() @ MyError::Unauthorized
    )]
    pub user_balance: Account<'info, UserBalance>,

    #[account(mut)]
    pub user: Signer<'info>,

    /// CHECK: This is the sysvar for instructions, used for Ed25519 signature verification.
    /// We are checking the account's address against the official sysvar ID, which is a sufficient safety check.
    #[account(address = anchor_lang::solana_program::sysvar::instructions::ID)]
    pub instructions: UncheckedAccount<'info>,

    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
#[instruction(batch_id: u64)]
pub struct CommitBatch<'info> {
    #[account(mut)]
    pub casino_vault: Account<'info, CasinoVault>,

    #[account(
        init,
        payer = authority,
        space = 8 + 32 + 8 + 32,
        seeds = [b"batch_commit", batch_id.to_le_bytes().as_ref()],
        bump
    )]
    pub batch_commit: Account<'info, BatchCommit>,

    #[account(mut)]
    pub authority: Signer<'info>,

    pub system_program: Program<'info, System>,
}

#[error_code]
pub enum MyError {
    #[msg("You are not authorized to perform this action.")]
    Unauthorized,
    #[msg("Signature verification failed. The provided server signature is invalid.")]
    SignatureVerificationFailed,
    #[msg("The casino vault has insufficient funds for this withdrawal.")]
    InsufficientVaultBalance,
    #[msg("The user's on-chain balance is insufficient for this withdrawal.")]
    InsufficientUserBalance,
    #[msg("The provided nonce is invalid or has already been used.")]
    InvalidNonce,
    #[msg("Deposit amount would cause an overflow.")]
    DepositOverflow,
}