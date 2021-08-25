const { default: Neon, rpc, wallet, tx, u,sc } = require("@cityofzion/neon-js");

//Neo RPC Node
const rpcClient = new rpc.RPCClient("http://seed1.ngd.network:10332");

const addressA = "AHckxqDYxMNGZydjj3DbGX3Y9vgfgVUafH";
const gasAssetId =
"602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7";

//CGAS hash(BigEnd scripthash)
const cgasHash = "74f2dc36a68fdc4682034178eb2220729231db76";
const fromAddress = "1436edb064beb3d6d9746fdf3ca3501cc9d8f599";//littleEnd scripthash
const toAddress = "4b721e06b50cc74e68b417716e3b099fb99757a8";//BlackHole Address: ANeo2toNeo3MigrationAddressxwPB2Hz (littleEnd scripthash)
const amount = "2000000000";//20 CGAS
//var from_Address = sc.ContractParam.byteArray(fromAddress, "address"); 
const from_Address = Neon.create.contractParam("ByteArray", fromAddress);
// var to_Address = sc.ContractParam.byteArray(toAddress, "address"); 
const to_Address = Neon.create.contractParam("ByteArray", toAddress);

//your privatekey
const wif = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"; 
const sigAddress = new wallet.Account(wif);
//UTXO hash
const input = "78334b2a1b04f80b32238154ddd0fe90f5ac17f62d1047ddcbcc229db5f56aed";


sendCCtx();

// create transfer transaction
  async function sendCCtx() {

    var sb = Neon.create.scriptBuilder();
    sb.emitAppCall(cgasHash, "transfer", [from_Address, to_Address, amount]);
    var script = sb.str;
    //create script
    let invocationTx = Neon.create.invocationTx();
    invocationTx.script = script
    invocationTx.fees = 0;

    let inputObj = {
      prevHash: input, // txid, utxo
      prevIndex: 0
    };

    let outPutObj = {
      assetId: gasAssetId, //change address
      value: 0.01, //output = input, just make utxo into tx, no need to spend
      scriptHash: wallet.getScriptHashFromAddress(addressA)  //big ending
    };

    // add transaction inputs and outputs
    invocationTx.inputs[0] = new tx.TransactionInput(inputObj);
    invocationTx.addOutput(new tx.TransactionOutput(outPutObj));
    invocationTx.addAttribute(
      tx.TxAttrUsage.Remark14,
      u.str2hexstring("Your N3 address")
    );
    // sign transaction with sender's private key
    const signature = wallet.sign(
      invocationTx.serialize(false), sigAddress.privateKey
    );

    // // add witness
    invocationTx.addWitness(
    tx.Witness.fromSignature(signature, sigAddress.publicKey)
    );

    console.log("send amount: " + amount)
    console.log("invocationTx: " + invocationTx.serialize())
    console.log("sending tx: ",invocationTx.hash)
    rpcClient
        .sendRawTransaction(invocationTx)
        .then(response => {
          console.log("success? ",response);
        })
        .catch(err => {
          console.log(err);
        });
  }
