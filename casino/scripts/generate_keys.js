const { randomBytes } = require("crypto");
const secp256k1 = require("secp256k1");
const { keccak_256 } = require("js-sha3");

let privKey;
do {
    privKey = randomBytes(32);
} while (!secp256k1.privateKeyVerify(privKey));

const pubKey = secp256k1.publicKeyCreate(privKey, false);
const authorityHash = Buffer.from(keccak_256.digest(pubKey.slice(1)));

console.log("--- SAVE THESE SECURELY (For your .env file) ---");
console.log("SERVER_SECP_PRIVATE_KEY_HEX:", privKey.toString('hex'));
console.log("------------------------------------------------");
console.log("--- USE THIS FOR INITIALIZE SCRIPT ---");
console.log("AUTHORITY_HASH_ARRAY:", JSON.stringify(Array.from(authorityHash)));