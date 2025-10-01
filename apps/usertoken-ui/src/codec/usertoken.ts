import type { GeneratedType } from "@cosmjs/proto-signing";
import { Reader, Writer } from "protobufjs/minimal";

export interface MsgCreateUserTokenProto {
  creator: string;
  subdenom: string;
  name: string;
  symbol: string;
  decimals: number;
}

export interface MsgBuyFounderTokensProto {
  buyer: string;
  denom: string;
}

export interface MsgBuyTokensProto {
  buyer: string;
  denom: string;
  amount: {
    denom: string;
    amount: string;
  };
  min_tokens: string;
}

export interface MsgSellTokensProto {
  seller: string;
  denom: string;
  amount: {
    denom: string;
    amount: string;
  };
  min_price: string;
}

type GeneratedTypeWithPartial<T> = GeneratedType & {
  fromPartial(object: Partial<T>): T;
};

const defaultCreateTokenMsg: MsgCreateUserTokenProto = {
  creator: "",
  subdenom: "",
  name: "",
  symbol: "",
  decimals: 0
};

const defaultBuyFounderMsg: MsgBuyFounderTokensProto = {
  buyer: "",
  denom: ""
};

const defaultBuyTokensMsg: MsgBuyTokensProto = {
  buyer: "",
  denom: "",
  amount: {
    denom: "",
    amount: ""
  },
  min_tokens: ""
};

const defaultSellTokensMsg: MsgSellTokensProto = {
  seller: "",
  denom: "",
  amount: {
    denom: "",
    amount: ""
  },
  min_price: ""
};

export const MSG_CREATE_USER_TOKEN_TYPE_URL = "/osmosis.usertoken.v1beta1.MsgCreateUserToken";
export const MSG_BUY_FOUNDER_TOKENS_TYPE_URL = "/osmosis.usertoken.v1beta1.MsgBuyFounderTokens";
export const MSG_BUY_TOKENS_TYPE_URL = "/osmosis.usertoken.v1beta1.MsgBuyTokens";
export const MSG_SELL_TOKENS_TYPE_URL = "/osmosis.usertoken.v1beta1.MsgSellTokens";

export const MsgCreateUserToken: GeneratedTypeWithPartial<MsgCreateUserTokenProto> = {
  encode(message: MsgCreateUserTokenProto, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.subdenom !== "") {
      writer.uint32(18).string(message.subdenom);
    }
    if (message.name !== "") {
      writer.uint32(26).string(message.name);
    }
    if (message.symbol !== "") {
      writer.uint32(34).string(message.symbol);
    }
    if (message.decimals !== 0) {
      writer.uint32(40).uint32(message.decimals);
    }
    return writer;
  },

  decode(input: Uint8Array | Reader, length?: number): MsgCreateUserTokenProto {
    const reader = input instanceof Reader ? input : new Reader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message: MsgCreateUserTokenProto = { ...defaultCreateTokenMsg };

    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.subdenom = reader.string();
          break;
        case 3:
          message.name = reader.string();
          break;
        case 4:
          message.symbol = reader.string();
          break;
        case 5:
          message.decimals = reader.uint32();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }

    return message;
  },

  fromPartial(object: Partial<MsgCreateUserTokenProto>): MsgCreateUserTokenProto {
    return {
      creator: object.creator ?? "",
      subdenom: object.subdenom ?? "",
      name: object.name ?? "",
      symbol: object.symbol ?? "",
      decimals: object.decimals ?? 0
    };
  }
};

export const MsgBuyFounderTokens: GeneratedTypeWithPartial<MsgBuyFounderTokensProto> = {
  encode(message: MsgBuyFounderTokensProto, writer: Writer = Writer.create()): Writer {
    if (message.buyer !== "") {
      writer.uint32(10).string(message.buyer);
    }
    if (message.denom !== "") {
      writer.uint32(18).string(message.denom);
    }
    return writer;
  },

  decode(input: Uint8Array | Reader, length?: number): MsgBuyFounderTokensProto {
    const reader = input instanceof Reader ? input : new Reader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message: MsgBuyFounderTokensProto = { ...defaultBuyFounderMsg };

    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.buyer = reader.string();
          break;
        case 2:
          message.denom = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }

    return message;
  },

  fromPartial(object: Partial<MsgBuyFounderTokensProto>): MsgBuyFounderTokensProto {
    return {
      buyer: object.buyer ?? "",
      denom: object.denom ?? ""
    };
  }
};

export const MsgBuyTokens: GeneratedTypeWithPartial<MsgBuyTokensProto> = {
  encode(message: MsgBuyTokensProto, writer: Writer = Writer.create()): Writer {
    if (message.buyer !== "") {
      writer.uint32(10).string(message.buyer);
    }
    if (message.denom !== "") {
      writer.uint32(18).string(message.denom);
    }
    if (message.amount.denom !== "" || message.amount.amount !== "") {
      writer.uint32(26).fork();
      if (message.amount.denom !== "") {
        writer.uint32(10).string(message.amount.denom);
      }
      if (message.amount.amount !== "") {
        writer.uint32(18).string(message.amount.amount);
      }
      writer.ldelim();
    }
    if (message.min_tokens !== "") {
      writer.uint32(34).string(message.min_tokens);
    }
    return writer;
  },

  decode(input: Uint8Array | Reader, length?: number): MsgBuyTokensProto {
    const reader = input instanceof Reader ? input : new Reader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message: MsgBuyTokensProto = { ...defaultBuyTokensMsg };

    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.buyer = reader.string();
          break;
        case 2:
          message.denom = reader.string();
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
        case 4:
          message.min_tokens = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }

    return message;
  },

  fromPartial(object: Partial<MsgBuyTokensProto>): MsgBuyTokensProto {
    return {
      buyer: object.buyer ?? "",
      denom: object.denom ?? "",
      amount: {
        denom: object.amount?.denom ?? "",
        amount: object.amount?.amount ?? ""
      },
      min_tokens: object.min_tokens ?? ""
    };
  }
};

export const MsgSellTokens: GeneratedTypeWithPartial<MsgSellTokensProto> = {
  encode(message: MsgSellTokensProto, writer: Writer = Writer.create()): Writer {
    if (message.seller !== "") {
      writer.uint32(10).string(message.seller);
    }
    if (message.denom !== "") {
      writer.uint32(18).string(message.denom);
    }
    if (message.amount.denom !== "" || message.amount.amount !== "") {
      writer.uint32(26).fork();
      if (message.amount.denom !== "") {
        writer.uint32(10).string(message.amount.denom);
      }
      if (message.amount.amount !== "") {
        writer.uint32(18).string(message.amount.amount);
      }
      writer.ldelim();
    }
    if (message.min_price !== "") {
      writer.uint32(34).string(message.min_price);
    }
    return writer;
  },

  decode(input: Uint8Array | Reader, length?: number): MsgSellTokensProto {
    const reader = input instanceof Reader ? input : new Reader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message: MsgSellTokensProto = { ...defaultSellTokensMsg };

    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.seller = reader.string();
          break;
        case 2:
          message.denom = reader.string();
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
        case 4:
          message.min_price = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }

    return message;
  },

  fromPartial(object: Partial<MsgSellTokensProto>): MsgSellTokensProto {
    return {
      seller: object.seller ?? "",
      denom: object.denom ?? "",
      amount: {
        denom: object.amount?.denom ?? "",
        amount: object.amount?.amount ?? ""
      },
      min_price: object.min_price ?? ""
    };
  }
};
