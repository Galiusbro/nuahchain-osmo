// Protobuf message types for leverage module
// These should match the protobuf definitions in the backend

import { GeneratedType } from "@cosmjs/proto-signing";
import { Reader, Writer } from "protobufjs/minimal";

// Type definition
type GeneratedTypeWithPartial<T> = GeneratedType & {
    fromPartial(object: Partial<T>): T;
};

// Message interfaces
export interface MsgOpenPositionProto {
    trader: string;
    token_denom: string;
    collateral: {
        denom: string;
        amount: string;
    };
    leverage: string;
    side: number;
    min_price: string;
    max_price: string;
}

export interface MsgClosePositionProto {
    trader: string;
    position_id: string;
    min_price: string;
    max_price: string;
}

export interface MsgAddCollateralProto {
    trader: string;
    position_id: string;
    amount: {
        denom: string;
        amount: string;
    };
}

export interface MsgRemoveCollateralProto {
    trader: string;
    position_id: string;
    amount: {
        denom: string;
        amount: string;
    };
}

export interface MsgLiquidatePositionProto {
    liquidator: string;
    position_id: string;
}

export interface MsgProvideLiquidityProto {
    provider: string;
    amount: {
        denom: string;
        amount: string;
    };
}

// Simple message types - these will be handled by the protobuf library
export const MsgOpenPosition: GeneratedTypeWithPartial<MsgOpenPositionProto> = {
    encode: (message: MsgOpenPositionProto, writer: Writer = Writer.create()): Writer => {
        if (message.trader !== "") {
            writer.uint32(10).string(message.trader);
        }
        if (message.token_denom !== "") {
            writer.uint32(18).string(message.token_denom);
        }
        if (message.collateral !== undefined) {
            writer.uint32(26).fork();
            writer.uint32(10).string(message.collateral.denom);
            writer.uint32(18).string(message.collateral.amount);
            writer.ldelim();
        }
        if (message.leverage !== "") {
            writer.uint32(34).string(message.leverage);
        }
        if (message.side !== 0) {
            writer.uint32(40).int32(message.side);
        }
        if (message.min_price !== "") {
            writer.uint32(50).string(message.min_price);
        }
        if (message.max_price !== "") {
            writer.uint32(58).string(message.max_price);
        }
        return writer;
    },
    decode: (input: Uint8Array | Reader, length?: number): MsgOpenPositionProto => {
        const reader = input instanceof Reader ? input : new Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseMsgOpenPositionProto();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.trader = reader.string();
                    break;
                case 2:
                    message.token_denom = reader.string();
                    break;
                case 3:
                    const coinEnd = reader.uint32() + reader.pos;
                    message.collateral = { denom: "", amount: "" };
                    while (reader.pos < coinEnd) {
                        const coinTag = reader.uint32();
                        switch (coinTag >>> 3) {
                            case 1:
                                message.collateral.denom = reader.string();
                                break;
                            case 2:
                                message.collateral.amount = reader.string();
                                break;
                            default:
                                reader.skipType(coinTag & 7);
                                break;
                        }
                    }
                    break;
                case 4:
                    message.leverage = reader.string();
                    break;
                case 5:
                    message.side = reader.int32();
                    break;
                case 6:
                    message.min_price = reader.string();
                    break;
                case 7:
                    message.max_price = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    create: (properties?: { [k: string]: any }): MsgOpenPositionProto => {
        const message = properties as Partial<MsgOpenPositionProto> || {};
        return {
            trader: message.trader ?? "",
            token_denom: message.token_denom ?? "",
            collateral: message.collateral ?? { denom: "", amount: "" },
            leverage: message.leverage ?? "",
            side: message.side ?? 0,
            min_price: message.min_price ?? "",
            max_price: message.max_price ?? ""
        };
    },
    fromPartial: (object: Partial<MsgOpenPositionProto>): MsgOpenPositionProto => {
        return {
            trader: object.trader ?? "",
            token_denom: object.token_denom ?? "",
            collateral: object.collateral ?? { denom: "", amount: "" },
            leverage: object.leverage ?? "",
            side: object.side ?? 0,
            min_price: object.min_price ?? "",
            max_price: object.max_price ?? ""
        };
    }
};

function createBaseMsgOpenPositionProto(): MsgOpenPositionProto {
    return {
        trader: "",
        token_denom: "",
        collateral: { denom: "", amount: "" },
        leverage: "",
        side: 0,
        min_price: "",
        max_price: ""
    };
}

export const MsgClosePosition: GeneratedTypeWithPartial<MsgClosePositionProto> = {
    encode: (message: MsgClosePositionProto, writer: Writer = Writer.create()): Writer => {
        if (message.trader !== "") {
            writer.uint32(10).string(message.trader);
        }
        if (message.position_id !== "") {
            writer.uint32(18).string(message.position_id);
        }
        if (message.min_price !== "") {
            writer.uint32(26).string(message.min_price);
        }
        if (message.max_price !== "") {
            writer.uint32(34).string(message.max_price);
        }
        return writer;
    },
    decode: (input: Uint8Array | Reader, length?: number): MsgClosePositionProto => {
        const reader = input instanceof Reader ? input : new Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseMsgClosePositionProto();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.trader = reader.string();
                    break;
                case 2:
                    message.position_id = reader.string();
                    break;
                case 3:
                    message.min_price = reader.string();
                    break;
                case 4:
                    message.max_price = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    create: (properties?: { [k: string]: any }): MsgClosePositionProto => {
        const message = properties as Partial<MsgClosePositionProto> || {};
        return {
            trader: message.trader ?? "",
            position_id: message.position_id ?? "",
            min_price: message.min_price ?? "",
            max_price: message.max_price ?? ""
        };
    },
    fromPartial: (object: Partial<MsgClosePositionProto>): MsgClosePositionProto => {
        return {
            trader: object.trader ?? "",
            position_id: object.position_id ?? "",
            min_price: object.min_price ?? "",
            max_price: object.max_price ?? ""
        };
    }
};

function createBaseMsgClosePositionProto(): MsgClosePositionProto {
    return {
        trader: "",
        position_id: "",
        min_price: "",
        max_price: ""
    };
}

export const MsgAddCollateral: GeneratedTypeWithPartial<MsgAddCollateralProto> = {
    encode: (message: MsgAddCollateralProto, writer: Writer = Writer.create()): Writer => {
        if (message.trader !== "") {
            writer.uint32(10).string(message.trader);
        }
        if (message.position_id !== "") {
            writer.uint32(18).string(message.position_id);
        }
        if (message.amount !== undefined) {
            writer.uint32(26).fork();
            writer.uint32(10).string(message.amount.denom);
            writer.uint32(18).string(message.amount.amount);
            writer.ldelim();
        }
        return writer;
    },
    decode: (input: Uint8Array | Reader, length?: number): MsgAddCollateralProto => {
        const reader = input instanceof Reader ? input : new Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseMsgAddCollateralProto();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.trader = reader.string();
                    break;
                case 2:
                    message.position_id = reader.string();
                    break;
                case 3:
                    const coinEnd = reader.uint32() + reader.pos;
                    while (reader.pos < coinEnd) {
                        const coinTag = reader.uint32();
                        switch (coinTag >>> 3) {
                            case 1:
                                message.amount.denom = reader.string();
                                break;
                            case 2:
                                message.amount.amount = reader.string();
                                break;
                            default:
                                reader.skipType(coinTag & 7);
                                break;
                        }
                    }
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    create: (properties?: { [k: string]: any }): MsgAddCollateralProto => {
        const message = properties as Partial<MsgAddCollateralProto> || {};
        return {
            trader: message.trader ?? "",
            position_id: message.position_id ?? "",
            amount: message.amount ?? { denom: "", amount: "" }
        };
    },
    fromPartial: (object: Partial<MsgAddCollateralProto>): MsgAddCollateralProto => {
        return {
            trader: object.trader ?? "",
            position_id: object.position_id ?? "",
            amount: object.amount ?? { denom: "", amount: "" }
        };
    }
};

function createBaseMsgAddCollateralProto(): MsgAddCollateralProto {
    return {
        trader: "",
        position_id: "",
        amount: { denom: "", amount: "" }
    };
}

export const MsgRemoveCollateral: GeneratedTypeWithPartial<MsgRemoveCollateralProto> = {
    encode: (message: MsgRemoveCollateralProto, writer: Writer = Writer.create()): Writer => {
        if (message.trader !== "") {
            writer.uint32(10).string(message.trader);
        }
        if (message.position_id !== "") {
            writer.uint32(18).string(message.position_id);
        }
        if (message.amount !== undefined) {
            writer.uint32(26).fork();
            writer.uint32(10).string(message.amount.denom);
            writer.uint32(18).string(message.amount.amount);
            writer.ldelim();
        }
        return writer;
    },
    decode: (input: Uint8Array | Reader, length?: number): MsgRemoveCollateralProto => {
        const reader = input instanceof Reader ? input : new Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseMsgRemoveCollateralProto();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.trader = reader.string();
                    break;
                case 2:
                    message.position_id = reader.string();
                    break;
                case 3:
                    const coinEnd = reader.uint32() + reader.pos;
                    while (reader.pos < coinEnd) {
                        const coinTag = reader.uint32();
                        switch (coinTag >>> 3) {
                            case 1:
                                message.amount.denom = reader.string();
                                break;
                            case 2:
                                message.amount.amount = reader.string();
                                break;
                            default:
                                reader.skipType(coinTag & 7);
                                break;
                        }
                    }
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    create: (properties?: { [k: string]: any }): MsgRemoveCollateralProto => {
        const message = properties as Partial<MsgRemoveCollateralProto> || {};
        return {
            trader: message.trader ?? "",
            position_id: message.position_id ?? "",
            amount: message.amount ?? { denom: "", amount: "" }
        };
    },
    fromPartial: (object: Partial<MsgRemoveCollateralProto>): MsgRemoveCollateralProto => {
        return {
            trader: object.trader ?? "",
            position_id: object.position_id ?? "",
            amount: object.amount ?? { denom: "", amount: "" }
        };
    }
};

function createBaseMsgRemoveCollateralProto(): MsgRemoveCollateralProto {
    return {
        trader: "",
        position_id: "",
        amount: { denom: "", amount: "" }
    };
}

export const MsgLiquidatePosition: GeneratedTypeWithPartial<MsgLiquidatePositionProto> = {
    encode: (message: MsgLiquidatePositionProto, writer: Writer = Writer.create()): Writer => {
        if (message.liquidator !== "") {
            writer.uint32(10).string(message.liquidator);
        }
        if (message.position_id !== "") {
            writer.uint32(18).string(message.position_id);
        }
        return writer;
    },
    decode: (input: Uint8Array | Reader, length?: number): MsgLiquidatePositionProto => {
        const reader = input instanceof Reader ? input : new Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseMsgLiquidatePositionProto();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.liquidator = reader.string();
                    break;
                case 2:
                    message.position_id = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    create: (properties?: { [k: string]: any }): MsgLiquidatePositionProto => {
        const message = properties as Partial<MsgLiquidatePositionProto> || {};
        return {
            liquidator: message.liquidator ?? "",
            position_id: message.position_id ?? ""
        };
    },
    fromPartial: (object: Partial<MsgLiquidatePositionProto>): MsgLiquidatePositionProto => {
        return {
            liquidator: object.liquidator ?? "",
            position_id: object.position_id ?? ""
        };
    }
};

function createBaseMsgLiquidatePositionProto(): MsgLiquidatePositionProto {
    return {
        liquidator: "",
        position_id: ""
    };
}

export const MsgProvideLiquidity: GeneratedTypeWithPartial<MsgProvideLiquidityProto> = {
    encode: (message: MsgProvideLiquidityProto, writer: Writer = Writer.create()): Writer => {
        if (message.provider !== "") {
            writer.uint32(10).string(message.provider);
        }
        if (message.amount !== undefined) {
            writer.uint32(18).fork();
            writer.uint32(10).string(message.amount.denom);
            writer.uint32(18).string(message.amount.amount);
            writer.ldelim();
        }
        return writer;
    },
    decode: (input: Uint8Array | Reader, length?: number): MsgProvideLiquidityProto => {
        const reader = input instanceof Reader ? input : new Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = createBaseMsgProvideLiquidityProto();
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.provider = reader.string();
                    break;
                case 2:
                    const coinEnd = reader.uint32() + reader.pos;
                    while (reader.pos < coinEnd) {
                        const coinTag = reader.uint32();
                        switch (coinTag >>> 3) {
                            case 1:
                                message.amount.denom = reader.string();
                                break;
                            case 2:
                                message.amount.amount = reader.string();
                                break;
                            default:
                                reader.skipType(coinTag & 7);
                                break;
                        }
                    }
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    create: (properties?: { [k: string]: any }): MsgProvideLiquidityProto => {
        const message = properties as Partial<MsgProvideLiquidityProto> || {};
        return {
            provider: message.provider ?? "",
            amount: message.amount ?? { denom: "", amount: "" }
        };
    },
    fromPartial: (object: Partial<MsgProvideLiquidityProto>): MsgProvideLiquidityProto => {
        return {
            provider: object.provider ?? "",
            amount: object.amount ?? { denom: "", amount: "" }
        };
    }
};

function createBaseMsgProvideLiquidityProto(): MsgProvideLiquidityProto {
    return {
        provider: "",
        amount: { denom: "", amount: "" }
    };
}
