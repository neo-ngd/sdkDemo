package io.neow3j.examples.jsonrpc;

import io.neow3j.crypto.transaction.RawTransactionAttribute;
import io.neow3j.model.types.NEOAsset;
import io.neow3j.model.types.TransactionAttributeUsageType;
import io.neow3j.protocol.Neow3j;
import io.neow3j.protocol.exceptions.ErrorResponseException;
import io.neow3j.protocol.http.HttpService;
import io.neow3j.wallet.*;
import io.neow3j.contract.Nep5;
import io.neow3j.contract.ScriptHash;
import io.neow3j.model.types.NEOAsset;
import io.neow3j.protocol.Neow3j;
import io.neow3j.protocol.exceptions.ErrorResponseException;
import io.neow3j.protocol.http.HttpService;
import io.neow3j.wallet.Account;
import io.neow3j.wallet.AssetTransfer;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.LinkedList;
import java.util.List;

public class TestUtxoTx {
    public static void main(String[] args) throws IOException, ErrorResponseException {
        Neow3j neow3j = Neow3j.build(new HttpService("http://seed1.ngd.network:20332"));
        String N3Address = "NiRtFpjqsAUrcpfGsqmj3uaUhtgM8SFLf9";
        Account useraccount = Account.fromWIF("wif string").build();
        useraccount.updateAssetBalances(neow3j);
        byte[] data = useraccount.getScriptHash().toArray();
        RawTransactionAttribute attribute1 = new RawTransactionAttribute(TransactionAttributeUsageType.SCRIPT, data);
        RawTransactionAttribute attribute2 = new RawTransactionAttribute(TransactionAttributeUsageType.REMARK1, new byte[32]);

        RawTransactionAttribute attribute3 = new RawTransactionAttribute(TransactionAttributeUsageType.REMARK14, N3Address.getBytes(StandardCharsets.UTF_8));


        List<RawTransactionAttribute> attributes = new LinkedList<>();
        attributes.add(attribute1);
        attributes.add(attribute2);
        attributes.add(attribute3);

        AssetTransfer transfer = new AssetTransfer.Builder(neow3j)
                .account(useraccount)
                .output(NEOAsset.HASH_ID, 1, "AJ36ZCpMhiHYMdMAUaP7i1i9pJz4jMdiQV")
                .attributes(attributes)
                .networkFee(1)
                .build()
                .sign()
                .send();
        System.out.println("\n####################");
        System.out.println("Transfer done.");
        System.out.println("txid: "+transfer.getTransaction().getTxId());
        System.out.println("####################\n");

    }
}

