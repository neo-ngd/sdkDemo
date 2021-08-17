# Migration Golang Demo 

---

Make sure you read and understand the code. 
Do not just copy and paste the code.
Use the code as a reference and write your own migration code. 
And the author is not responsible for any asset loss.

---

Before you get started, make sure you have the right black hole N3 address. 
And a stable node rpc port is also required.

Then a neo wallet file in JSON format or some neo accounts are needed. 
The demo code can even handle the situation that you have multiple accounts, 
and balance of each account is less than the gate value, 
but the total balance of all accounts are equal to/greater than the gate value. 
If you only have one account, the code could be much simpler.

Next, you just need to specify the asset hash and the amount you want to migrate. 
Also, a valid N3 address is a must-have.

Finally, use the correct method according to asset type (utxo or nep5) and happy migrating!
