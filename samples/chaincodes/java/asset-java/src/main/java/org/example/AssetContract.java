package org.example;

import com.google.gson.Gson;
import org.hyperledger.fabric.contract.Context;
import org.hyperledger.fabric.contract.ContractInterface;
import org.hyperledger.fabric.contract.annotation.Contract;
import org.hyperledger.fabric.contract.annotation.Default;
import org.hyperledger.fabric.contract.annotation.Transaction;

@Contract(name = "AssetContract")
@Default
public class AssetContract implements ContractInterface {
    private final Gson gson = new Gson();

    public static class Asset {
        public String id;
        public String owner;
        public int value;

        public Asset(String id, String owner, int value) {
            this.id = id;
            this.owner = owner;
            this.value = value;
        }
    }

    @Transaction(intent = Transaction.TYPE.SUBMIT)
    public void CreateAsset(Context ctx, String id, String owner, int value) {
        String existing = ctx.getStub().getStringState(id);
        if (existing != null && !existing.isEmpty()) {
            throw new IllegalArgumentException("asset " + id + " already exists");
        }
        ctx.getStub().putStringState(id, gson.toJson(new Asset(id, owner, value)));
    }

    @Transaction(intent = Transaction.TYPE.EVALUATE)
    public String ReadAsset(Context ctx, String id) {
        String state = ctx.getStub().getStringState(id);
        if (state == null || state.isEmpty()) {
            throw new IllegalArgumentException("asset " + id + " not found");
        }
        return state;
    }

    @Transaction(intent = Transaction.TYPE.SUBMIT)
    public void TransferAsset(Context ctx, String id, String newOwner) {
        Asset asset = gson.fromJson(ReadAsset(ctx, id), Asset.class);
        asset.owner = newOwner;
        ctx.getStub().putStringState(id, gson.toJson(asset));
    }
}
