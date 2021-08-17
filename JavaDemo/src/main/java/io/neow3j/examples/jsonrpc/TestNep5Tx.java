package io.neow3j.examples.jsonrpc;
import io.neow3j.contract.ContractInvocation;
import io.neow3j.contract.ContractParameter;

import io.neow3j.contract.ScriptHash;
import io.neow3j.crypto.transaction.RawTransactionAttribute;

import io.neow3j.model.types.TransactionAttributeUsageType;
import io.neow3j.protocol.Neow3j;
import io.neow3j.protocol.exceptions.ErrorResponseException;
import io.neow3j.protocol.http.HttpService;
import io.neow3j.wallet.Account;


import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.LinkedList;
import java.util.List;

public class TestNep5Tx {

    private static final String N3Address = "wif string";


    public static void main(String[] args) throws IOException, ErrorResponseException {
        Neow3j neow3j = Neow3j.build(new HttpService("http://seed1.ngd.network:20332"));

        Account account1 = Account.fromWIF("Kywf14wnQXgEodpnXEDgpiEWaGSDRn2TFeLekaDBn9DP28kPbYCd").build();
        account1.updateAssetBalances(neow3j);

        Account account2 = Account.fromAddress("AJ36ZCpMhiHYMdMAUaP7i1i9pJz4jMdiQV").build();
        //Account account2 = Account.fromWIF("KwbsvNed8G1HuA3riumvFtpbJmYWmiQkGuH3DczqqwnczF4qST5B").build();
        account2.updateAssetBalances(neow3j);



        ScriptHash contractScripthash = new ScriptHash("0x74f2dc36a68fdc4682034178eb2220729231db76");


        List<ContractParameter> params = new ArrayList<>();
        params.add(ContractParameter.byteArrayFromAddress(account1.getAddress()));
        params.add(ContractParameter.byteArrayFromAddress(account2.getScriptHash().toAddress()));
        params.add(ContractParameter.integer(2000000001));

        byte[] data = account1.getScriptHash().toArray();
        RawTransactionAttribute attribute1 = new RawTransactionAttribute(TransactionAttributeUsageType.SCRIPT, data);
        RawTransactionAttribute attribute2 = new RawTransactionAttribute(TransactionAttributeUsageType.REMARK1, new byte[32]);
        RawTransactionAttribute attribute3 = new RawTransactionAttribute(TransactionAttributeUsageType.REMARK14, N3Address.getBytes(StandardCharsets.UTF_8));
        List<RawTransactionAttribute> attributes = new LinkedList<>();
        attributes.add(attribute1);
        attributes.add(attribute2);
        attributes.add(attribute3);

       ContractInvocation c = new ContractInvocation.Builder(neow3j)
                .contractScriptHash(contractScripthash)
                .function("transfer")
                .parameters(params)
                .attributes(attributes)
                .account(account1)
                .networkFee(1)
                .build()
                .sign()
                .invoke();

        Boolean transferState =  c.getResponse().getResult();
        if (transferState) {
            System.out.println("\n####################");
            System.out.println("Transfer successful.");
            System.out.println("txid: "+c.getTransaction().getTxId());
            System.out.println("####################\n");
        }
    }

}
