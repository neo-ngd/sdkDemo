using Neo;
using Neo.IO;
using Neo.Ledger;
using Neo.Network.P2P.Payloads;
using Neo.SmartContract;
using Neo.VM;
using Neo.Wallets;
using Newtonsoft.Json.Linq;
using System;
using System.Collections.Generic;
using System.IO;
using System.Net;
using System.Numerics;
using System.Text;

namespace CSharpDemo
{
    class SendTrans
    {
        private static UInt160 toAddress = "ATEjGpdidGyrzSzpzoqRDkHbpGzVbkUNUU".ToScriptHash(); // 指定的迁移销毁地址
        private static string rpc = "http://seed1.ngd.network:20332"; //Testnet rpc url

        private static UInt256 NEOHash = Blockchain.GoverningToken.Hash;
        private static UInt256 GASHash = Blockchain.UtilityToken.Hash;

        //Mainnet Hash see https://github.com/neo-ngd/sdkDemo/contracthash.md
        //private static UInt160 nNEOHash = UInt160.Parse("0xf46719e2d16bf50cddcef9d4bbfece901f73cbb6");
        //private static UInt160 cGASHash = UInt160.Parse("0x74f2dc36a68fdc4682034178eb2220729231db76");

        //Testnet Hash see https://github.com/neo-ngd/sdkDemo/contracthash.md
        private static UInt160 nNEOHash = UInt160.Parse("0x17da3881ab2d050fea414c80b3fa8324d756f60e");
        private static UInt160 cGASHash = UInt160.Parse("0x74f2dc36a68fdc4682034178eb2220729231db76");

        //amount 是包含精度的 BigInteger
        public static void SendNep5Transaction(BigInteger amount)
        {
            byte[] prikey = Wallet.GetPrivateKeyFromWIF("KzRy5fJJvsJr37GmkhGZMb6LjsDcWt15CNwzDsYQH27Smdde4Fp6");
            KeyPair keyPair = new KeyPair(prikey);
            var user = Contract.CreateSignatureContract(keyPair.PublicKey);

            byte[] script;
            using (ScriptBuilder sb = new ScriptBuilder())
            {
                //sb.EmitAppCall(nNEOHash, "transfer", user.ScriptHash, toAddress, amount); //nNEO nNEO 数量为整数，如果有小数部分 N3 收到的 NEO 会做向下取整
                sb.EmitAppCall(cGASHash, "transfer", user.ScriptHash, toAddress, amount); //cGAS
                script = sb.ToArray();
            }

            var tx = new InvocationTransaction();
            //nNEO < 10 或 cGAS < 20 就收 1 GAS 网络费
            if (amount < 20_00000000)
            {
                tx = MakeTranWithFee(user.Address, 1, script);
            }
            else
            {
                tx.Version = 0;
                tx.Inputs = new CoinReference[] { };
                tx.Outputs = new TransactionOutput[] { };
                tx.Script = script;
            }

            Random random = new Random();
            var nonce = new byte[32];
            random.NextBytes(nonce);
            tx.Attributes = new TransactionAttribute[]
            {
                new TransactionAttribute() { Usage = TransactionAttributeUsage.Script, Data = user.ScriptHash.ToArray() },
                new TransactionAttribute() { Usage = TransactionAttributeUsage.Remark1, Data = nonce },
                new TransactionAttribute() { Usage = TransactionAttributeUsage.Remark14, Data = Encoding.UTF8.GetBytes("NUs2zy9vTpaf5oUu1AKqgXGAhDrnQHt3uq") } //N3 address
            };


            var signature = tx.Sign(keyPair);

            var sb1 = new ScriptBuilder();
            var invocationScript = sb1.EmitPush(signature).ToArray();
            var verificationScript = Contract.CreateSignatureRedeemScript(keyPair.PublicKey);
            tx.Witnesses = new[] { new Witness() { InvocationScript = invocationScript, VerificationScript = verificationScript } };

            string rawdata = tx.ToArray().ToHexString();
            string result = InvokeRpc(rpc, "sendrawtransaction", rawdata); //send transaction to rpc

            Console.WriteLine("txid: " + tx.Hash.ToString());
            Console.WriteLine(result.ToString());
        }


        public static void SendGASTransaction(decimal amount)
        {
            byte[] prikey = Wallet.GetPrivateKeyFromWIF("KzRy5fJJvsJr37GmkhGZMb6LjsDcWt15CNwzDsYQH27Smdde4Fp6");
            KeyPair keyPair = new KeyPair(prikey);
            var user = Contract.CreateSignatureContract(keyPair.PublicKey);

            List<UTXO> gasList = GetUTXOsByAddress(GASHash, user.Address);

            ContractTransaction tx = new ContractTransaction();
            tx.Version = 0;

            Random random = new Random();
            var nonce = new byte[32];
            random.NextBytes(nonce);
            tx.Attributes = new TransactionAttribute[]
            {
                new TransactionAttribute() { Usage = TransactionAttributeUsage.Script, Data = user.ScriptHash.ToArray() },
                new TransactionAttribute() { Usage = TransactionAttributeUsage.Remark1, Data = nonce },
                new TransactionAttribute() { Usage = TransactionAttributeUsage.Remark14, Data = Encoding.UTF8.GetBytes("NUs2zy9vTpaf5oUu1AKqgXGAhDrnQHt3uq") } //N3 address
            };

            //GAS < 20 就多收 1 GAS 网络费
            decimal net_fee = amount > 20 ? 0 : 1;
            decimal need_gas = amount + net_fee;

            gasList.Sort((a, b) =>
            {
                if (a.value > b.value)
                    return 1;
                else if (a.value < b.value)
                    return -1;
                else
                    return 0;
            });

            decimal all_input = decimal.Zero;
            List<CoinReference> inputs = new List<CoinReference>();

            for (int i = 0; i < gasList.Count; i++)
            {
                CoinReference coin = new CoinReference();
                coin.PrevHash = gasList[i].txid;
                coin.PrevIndex = (ushort)gasList[i].n;
                inputs.Add(coin);
                all_input += gasList[i].value;
                if (all_input >= need_gas)
                    break;
            }

            List<TransactionOutput> outputs = new List<TransactionOutput>();
            TransactionOutput output = new TransactionOutput();
            output.AssetId = GASHash;
            output.Value = Fixed8.FromDecimal(amount);
            output.ScriptHash = toAddress;
            outputs.Add(output);

            var change = all_input - need_gas;
            if (change > decimal.Zero)
            {
                TransactionOutput outputchange = new TransactionOutput();
                outputchange.AssetId = GASHash;
                outputchange.ScriptHash = user.ScriptHash;
                outputchange.Value = Fixed8.FromDecimal(change);
                outputs.Add(outputchange);
            }

            tx.Inputs = inputs.ToArray();
            tx.Outputs = outputs.ToArray();

            var signature = tx.Sign(keyPair);

            var sb1 = new ScriptBuilder();
            var invocationScript = sb1.EmitPush(signature).ToArray();
            var verificationScript = Contract.CreateSignatureRedeemScript(keyPair.PublicKey);
            tx.Witnesses = new[] { new Witness() { InvocationScript = invocationScript, VerificationScript = verificationScript } };

            string rawdata = tx.ToArray().ToHexString();
            string result = InvokeRpc(rpc, "sendrawtransaction", rawdata); //send transaction to rpc

            Console.WriteLine("txid: " + tx.Hash.ToString());
            Console.WriteLine(result.ToString());
        }

        //发送 NEO 时 amount 只能是整数
        public static void SendNEOTransaction(decimal amount)
        {
            byte[] prikey = Wallet.GetPrivateKeyFromWIF("KzRy5fJJvsJr37GmkhGZMb6LjsDcWt15CNwzDsYQH27Smdde4Fp6");
            KeyPair keyPair = new KeyPair(prikey);
            var user = Contract.CreateSignatureContract(keyPair.PublicKey);

            List<UTXO> neoList = GetUTXOsByAddress(NEOHash, user.Address);

            ContractTransaction tx = new ContractTransaction();
            tx.Version = 0;

            Random random = new Random();
            var nonce = new byte[32];
            random.NextBytes(nonce);
            tx.Attributes = new TransactionAttribute[]
            {
                new TransactionAttribute() { Usage = TransactionAttributeUsage.Script, Data = user.ScriptHash.ToArray() },
                new TransactionAttribute() { Usage = TransactionAttributeUsage.Remark1, Data = nonce },
                new TransactionAttribute() { Usage = TransactionAttributeUsage.Remark14, Data = Encoding.UTF8.GetBytes("NUs2zy9vTpaf5oUu1AKqgXGAhDrnQHt3uq") } //N3 address
            };

            neoList.Sort((a, b) =>
            {
                if (a.value > b.value)
                    return 1;
                else if (a.value < b.value)
                    return -1;
                else
                    return 0;
            });

            decimal all_input_neo = decimal.Zero;
            List<CoinReference> inputs = new List<CoinReference>();

            for (int i = 0; i < neoList.Count; i++)
            {
                CoinReference coin = new CoinReference();
                coin.PrevHash = neoList[i].txid;
                coin.PrevIndex = (ushort)neoList[i].n;
                inputs.Add(coin);
                all_input_neo += neoList[i].value;
                if (all_input_neo >= amount)
                    break;
            }

            List<TransactionOutput> outputs = new List<TransactionOutput>();
            TransactionOutput output = new TransactionOutput();
            output.AssetId = NEOHash;
            output.Value = Fixed8.FromDecimal(amount);
            output.ScriptHash = toAddress;
            outputs.Add(output);

            var change_neo = all_input_neo - amount;
            if (change_neo > decimal.Zero)
            {
                TransactionOutput outputchange = new TransactionOutput();
                outputchange.AssetId = NEOHash;
                outputchange.ScriptHash = user.ScriptHash;
                outputchange.Value = Fixed8.FromDecimal(change_neo);
                outputs.Add(outputchange);
            }

            //NEO < 10 就多收 1 GAS 网络费
            decimal net_fee = amount > 10 ? 0 : 1;
            if (net_fee > 0)
            {
                List<UTXO> gasList = GetUTXOsByAddress(GASHash, user.Address);
               
                gasList.Sort((a, b) =>
                {
                    if (a.value > b.value)
                        return 1;
                    else if (a.value < b.value)
                        return -1;
                    else
                        return 0;
                });

                decimal all_input_gas = decimal.Zero;               
                for (int i = 0; i < gasList.Count; i++)
                {
                    CoinReference coin = new CoinReference();
                    coin.PrevHash = gasList[i].txid;
                    coin.PrevIndex = (ushort)gasList[i].n;
                    inputs.Add(coin);
                    all_input_gas += gasList[i].value;
                    if (all_input_gas >= net_fee)
                        break;
                }              
                
                var change_gas = all_input_gas - net_fee;
                if (change_gas > decimal.Zero)
                {
                    TransactionOutput outputchange = new TransactionOutput();
                    outputchange.AssetId = GASHash;
                    outputchange.ScriptHash = user.ScriptHash;
                    outputchange.Value = Fixed8.FromDecimal(change_gas);
                    outputs.Add(outputchange);
                }
            }

            tx.Inputs = inputs.ToArray();
            tx.Outputs = outputs.ToArray();

            var signature = tx.Sign(keyPair);

            var sb1 = new ScriptBuilder();
            var invocationScript = sb1.EmitPush(signature).ToArray();
            var verificationScript = Contract.CreateSignatureRedeemScript(keyPair.PublicKey);
            tx.Witnesses = new[] { new Witness() { InvocationScript = invocationScript, VerificationScript = verificationScript } };

            string rawdata = tx.ToArray().ToHexString();
            string result = InvokeRpc(rpc, "sendrawtransaction", rawdata); //send transaction to rpc

            Console.WriteLine("txid: " + tx.Hash.ToString());
            Console.WriteLine(result.ToString());
        }

        private static InvocationTransaction MakeTranWithFee(string userAddr, decimal gas_consumed, byte[] script)
        {
            List<UTXO> gasList = GetUTXOsByAddress(GASHash, userAddr);

            var tx = new InvocationTransaction();
            tx.Attributes = new TransactionAttribute[] { };
            tx.Version = 0;
            tx.Inputs = new CoinReference[] { };
            tx.Outputs = new TransactionOutput[] { };
            tx.Witnesses = new Witness[] { };
            tx.Script = script;           

            gasList.Sort((a, b) =>
            {
                if (a.value > b.value)
                    return 1;
                else if (a.value < b.value)
                    return -1;
                else
                    return 0;
            });

            decimal all_input_gas = decimal.Zero;
            List<CoinReference> inputs = new List<CoinReference>();
            for (int i = 0; i < gasList.Count; i++)
            {
                CoinReference coin = new CoinReference();
                coin.PrevHash = gasList[i].txid;
                coin.PrevIndex = (ushort)gasList[i].n;
                inputs.Add(coin);
                all_input_gas += gasList[i].value;
                if (all_input_gas >= gas_consumed)
                    break;
            }

            if (all_input_gas >= gas_consumed)
            {
                List<TransactionOutput> list_outputs = new List<TransactionOutput>();                
                var change = all_input_gas - gas_consumed;
                if (change > decimal.Zero)
                {
                    TransactionOutput outputchange = new TransactionOutput();
                    outputchange.AssetId = GASHash;
                    outputchange.ScriptHash = userAddr.ToScriptHash();
                    outputchange.Value = Fixed8.FromDecimal(change);
                    list_outputs.Add(outputchange);
                }

                tx.Inputs = inputs.ToArray();
                tx.Outputs = list_outputs.ToArray();
            }
            else
            {
                throw new Exception("no enough money!");
            }

            return tx;
        }

        private static List<UTXO> GetUTXOsByAddress(UInt256 token, string address)
        {
            JObject response = JObject.Parse(HttpGet(rpc + "/?jsonrpc=2.0&id=1&method=getunspents&params=['" + address + "']"));
            JArray resJA = (JArray)response["result"]["balance"];

            List<UTXO> Utxos = new List<UTXO>();

            foreach (JObject jAsset in resJA)
            {
                var asset_hash = "0x" + jAsset["asset_hash"].ToString();
                if (asset_hash != token.ToString())
                    continue;
                var jUnspent = jAsset["unspent"] as JArray;

                foreach (JObject j in jUnspent)
                {
                    UTXO utxo = new UTXO(UInt256.Parse(j["txid"].ToString()), decimal.Parse(j["value"].ToString()), int.Parse(j["n"].ToString()));

                    Utxos.Add(utxo);
                }
            }
            return Utxos;
        }

        private static string InvokeRpc(string url, string method, string data)
        {
            string input = @"{
	            'jsonrpc': '2.0',
                'method': '&',
	            'params': ['#'],
	            'id': '1'
                }";

            input = input.Replace("&", method);
            input = input.Replace("#", data);

            string result = HttpPost(url, input);
            return result;
        }

        private static string HttpPost(string url, string data)
        {
            HttpWebRequest req = WebRequest.CreateHttp(new Uri(url));
            req.ContentType = "application/json;charset=utf-8";

            req.Method = "POST";
            //req.Accept = "text/xml,text/javascript";
            req.ContinueTimeout = 10000;

            byte[] postData = Encoding.UTF8.GetBytes(data);
            Stream reqStream = req.GetRequestStream();
            reqStream.Write(postData, 0, postData.Length);
            //reqStream.Dispose();

            HttpWebResponse rsp = (HttpWebResponse)req.GetResponse();
            string result = GetResponseAsString(rsp);

            return result;
        }

        private static string GetResponseAsString(HttpWebResponse rsp)
        {
            Stream stream = null;
            StreamReader reader = null;

            try
            {
                stream = rsp.GetResponseStream();
                reader = new StreamReader(stream, Encoding.UTF8);

                return reader.ReadToEnd();
            }
            finally
            {
                if (reader != null)
                    reader.Close();
                if (stream != null)
                    stream.Close();

            }
        }

        private static string HttpGet(string url)
        {
            WebClient wc = new WebClient();
            return wc.DownloadString(url);
        }
    }

    public class UTXO
    {
        public UInt256 txid;
        public int n;
        public decimal value;

        public UTXO(UInt256 _txid, decimal _value, int _n)
        {
            txid = _txid;
            value = _value;
            n = _n;
        }
    }
}
