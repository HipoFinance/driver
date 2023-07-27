# Driver
The driver for hTON: Hipo liquid staking protocol

## What does the **Driver** do?
The **Driver** helps users, to stake and unstake their coins more easily. Here is what happens:

### Staking:
1. When a user sends some TON coins with the comment `d`, the coins are received by the protocol, and it sends automatically a `save_coins` message
to the user's jetton wallet, called *j-wallet* from now on in this document.

2. The user's j-wallet keeps the request to ask the protocol for receiving the tokens. However, j-wallet cannot ask it immediately because there might be a validation
round in progress.

3. User should wait until the current validation round gets finished and then sends some fees again with the comment `s`. Of course this is not very pleasant
   for the user, but is required to receive his/her tokens.
   
The **Driver** does the 3rd action in background by sending a `stake_coins` message. So, users can stake their coins just by sending the first `d` message only.

### Unstaking
1. When a user sends some few fractions of TON (e.g. 0.2 TON) with the comment `w`, the fees are received by the protocol, and it sends automatically a
   `reserve_tokens` message to Hipo Treasury Wallet.

2. The user's j-wallet must ask the Hipo Treasury to burn the tokens and pay back the TON coins. However, similar to staking, j-wallet cannot ask it immediately
   because the tokens might be lent for the current validation round.

3. User should wait until the current validation round gets finished and then sends some fees again with the comment `u`, which is not very pleasant
   again.
   
Again, the **Driver** does the 3rd action in background by sending a `withdraw_tokens` message. So, users can unstake their coins just by sending the first `w` message only.

## And how?

**Driver** consists of three processes running simultaneously: *Extraction*, *Stake*, and *Unstake*

### Extraction:

Periodically looks through the network transactions and finds out all j-wallets having some 'staking' or 'unstaking' requests. It keeps the requests' information
on separate tables in a database. It also keeps some information about processed transactions, so that ignores the processed transactions in the next periodic turn.

### Stake:

Checks the *staking* requests from the database filled by the *Extraction* process, and sends the required messages if the stake request is not waiting for
the current validation round. If a message faces any error, it skips and will be retried a few times in the next coming turns.

### Unstake:

Checks the *unstaking* requests from the database filled by the *Extraction* process, and sends the required messages for each request if the Hipo Treasury
has enough coins to pay that request. This process sorts the requests ascending based on the requested tokens, so that pay the most number of requests with a
specified budget. If a message faces any error, it skips and will be retried a few times in the next coming turns.

## Configuration

The configuration is done using `config.json` file. Here are the configurable parameters:

- `service_db_uri`: PostgreSQL's database URL.
- `network`: The network on which the protocol is running. can be either `mainnet` or `testnet`.
- `treasury_address`: The address of the Treasury wallet in Base64URL format.
- `mnemonic`: The 24 phrase words of the driver's wallet. For example: `"subway under balance ..."` (replace the ... with the remaining word).
- `mnemonic_url`: The URL of the file containing the mnemonic. Only one of the `mnemonic` and `mnemonic_url` parameters must be specified.
- `extract_interval`, `stake_interval`, `unstake_interval`: The three intervals for running extraction, stake, and unstake processes respectively.
  For example `5s` as 5 seconds, or `1m` as 1 minute.
- `max_retry`: The number of retrying to send a message if it faces any error.
