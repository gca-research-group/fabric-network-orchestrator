'use strict';

const { Contract } = require('fabric-contract-api');

class AssetContract extends Contract {
    async CreateAsset(ctx, id, owner, value) {
        if (await ctx.stub.getState(id).then(data => data.length > 0)) {
            throw new Error(`asset ${id} already exists`);
        }
        await ctx.stub.putState(id, Buffer.from(JSON.stringify({ id, owner, value: Number(value) })));
    }

    async ReadAsset(ctx, id) {
        const data = await ctx.stub.getState(id);
        if (!data || data.length === 0) {
            throw new Error(`asset ${id} not found`);
        }
        return data.toString();
    }

    async TransferAsset(ctx, id, newOwner) {
        const asset = JSON.parse(await this.ReadAsset(ctx, id));
        asset.owner = newOwner;
        await ctx.stub.putState(id, Buffer.from(JSON.stringify(asset)));
    }
}

module.exports.contracts = [AssetContract];
