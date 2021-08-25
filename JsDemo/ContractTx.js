
const Neon = require("@cityofzion/neon-js");
console.log(`Hi, Let us begin the transaction`);

let raw = new Neon.tx.ContractTransaction();
let outPutObj1 = {
    "assetId": "c56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b",
    "value": "",
    "scriptHash": ""
}

let outPutObj2 = {
    "assetId": "602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7",
    "value": "",
    "scriptHash": ""
}
let inputObj = {
    "prevHash": "",
    "prevIndex": 0
}

let inputObj2 = {
    "prevHash": "",
    "prevIndex": 1
}

raw.addOutput(new Neon.tx.TransactionOutput(outPutObj1));
raw.inputs[0] = new Neon.tx.TransactionInput(inputObj);
raw.addAttribute(Neon.tx.TxAttrUsage.Remark14,"Your N3 address");

raw.addOutput(new Neon.tx.TransactionOutput(outPutObj2));
raw.inputs[1] = new Neon.tx.TransactionInput(inputObj2);

raw.sign("")
const rpcClient = new Neon.rpc.RPCClient("http://seed6.ngd.network:20332");
rpcClient.sendRawTransaction(raw).then((response)=>{
    console.log(response);
}).catch((err)=>{
    console.log(err);
});




