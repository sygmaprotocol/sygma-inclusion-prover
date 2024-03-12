// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package abi

const ExecutorABI = `[
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "bridge",
          "type": "address"
        },
        {
          "internalType": "address",
          "name": "accessControl",
          "type": "address"
        }
      ],
      "stateMutability": "nonpayable",
      "type": "constructor"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "sender",
          "type": "address"
        },
        {
          "internalType": "bytes4",
          "name": "funcSig",
          "type": "bytes4"
        }
      ],
      "name": "AccessNotAllowed",
      "type": "error"
    },
    {
      "inputs": [],
      "name": "BridgeIsPaused",
      "type": "error"
    },
    {
      "inputs": [],
      "name": "EmptyProposalsArray",
      "type": "error"
    },
    {
      "inputs": [
        {
          "internalType": "contract IStateRootStorage",
          "name": "stateRootStorage",
          "type": "address"
        },
        {
          "internalType": "bytes32",
          "name": "stateRoot",
          "type": "bytes32"
        }
      ],
      "name": "StateRootDoesNotMatch",
      "type": "error"
    },
    {
      "inputs": [
        {
          "internalType": "bytes32",
          "name": "transferHash",
          "type": "bytes32"
        }
      ],
      "name": "TransferHashDoesNotMatchSlotValue",
      "type": "error"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "internalType": "bytes",
          "name": "lowLevelData",
          "type": "bytes"
        },
        {
          "indexed": false,
          "internalType": "uint8",
          "name": "originDomainID",
          "type": "uint8"
        },
        {
          "indexed": false,
          "internalType": "uint64",
          "name": "depositNonce",
          "type": "uint64"
        }
      ],
      "name": "FailedHandlerExecution",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "internalType": "uint8",
          "name": "originDomainID",
          "type": "uint8"
        },
        {
          "indexed": false,
          "internalType": "address",
          "name": "newRouter",
          "type": "address"
        }
      ],
      "name": "FeeRouterChanged",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "internalType": "uint8",
          "name": "originDomainID",
          "type": "uint8"
        },
        {
          "indexed": false,
          "internalType": "uint64",
          "name": "depositNonce",
          "type": "uint64"
        },
        {
          "indexed": false,
          "internalType": "bytes",
          "name": "handlerResponse",
          "type": "bytes"
        }
      ],
      "name": "ProposalExecution",
      "type": "event"
    },
    {
      "inputs": [],
      "name": "_accessControl",
      "outputs": [
        {
          "internalType": "contract IAccessControlSegregator",
          "name": "",
          "type": "address"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [],
      "name": "_bridge",
      "outputs": [
        {
          "internalType": "contract IBridge",
          "name": "",
          "type": "address"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [],
      "name": "_domainID",
      "outputs": [
        {
          "internalType": "uint8",
          "name": "",
          "type": "uint8"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "uint8",
          "name": "",
          "type": "uint8"
        }
      ],
      "name": "_originDomainIDToRouter",
      "outputs": [
        {
          "internalType": "address",
          "name": "",
          "type": "address"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "uint8",
          "name": "",
          "type": "uint8"
        },
        {
          "internalType": "uint256",
          "name": "",
          "type": "uint256"
        }
      ],
      "name": "_securityModels",
      "outputs": [
        {
          "internalType": "address",
          "name": "",
          "type": "address"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "uint8",
          "name": "",
          "type": "uint8"
        }
      ],
      "name": "_slotIndexes",
      "outputs": [
        {
          "internalType": "uint8",
          "name": "",
          "type": "uint8"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "uint8",
          "name": "originDomainID",
          "type": "uint8"
        },
        {
          "internalType": "address",
          "name": "newRouter",
          "type": "address"
        }
      ],
      "name": "adminChangeRouter",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "uint8",
          "name": "originDomainID",
          "type": "uint8"
        },
        {
          "internalType": "uint8",
          "name": "slotIndex",
          "type": "uint8"
        }
      ],
      "name": "adminChangeSlotIndex",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "uint8",
          "name": "originDomainID",
          "type": "uint8"
        },
        {
          "internalType": "uint64",
          "name": "depositNonce",
          "type": "uint64"
        }
      ],
      "name": "adminMarkNonceAsUsed",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "uint8",
          "name": "securityModel",
          "type": "uint8"
        },
        {
          "internalType": "address[]",
          "name": "verifiersAddresses",
          "type": "address[]"
        }
      ],
      "name": "adminSetVerifiers",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "components": [
            {
              "internalType": "uint8",
              "name": "originDomainID",
              "type": "uint8"
            },
            {
              "internalType": "uint8",
              "name": "securityModel",
              "type": "uint8"
            },
            {
              "internalType": "uint64",
              "name": "depositNonce",
              "type": "uint64"
            },
            {
              "internalType": "bytes32",
              "name": "resourceID",
              "type": "bytes32"
            },
            {
              "internalType": "bytes",
              "name": "data",
              "type": "bytes"
            },
            {
              "internalType": "bytes[]",
              "name": "storageProof",
              "type": "bytes[]"
            }
          ],
          "internalType": "struct Executor.Proposal",
          "name": "proposal",
          "type": "tuple"
        },
        {
          "internalType": "bytes[]",
          "name": "accountProof",
          "type": "bytes[]"
        },
        {
          "internalType": "uint256",
          "name": "slot",
          "type": "uint256"
        }
      ],
      "name": "executeProposal",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "components": [
            {
              "internalType": "uint8",
              "name": "originDomainID",
              "type": "uint8"
            },
            {
              "internalType": "uint8",
              "name": "securityModel",
              "type": "uint8"
            },
            {
              "internalType": "uint64",
              "name": "depositNonce",
              "type": "uint64"
            },
            {
              "internalType": "bytes32",
              "name": "resourceID",
              "type": "bytes32"
            },
            {
              "internalType": "bytes",
              "name": "data",
              "type": "bytes"
            },
            {
              "internalType": "bytes[]",
              "name": "storageProof",
              "type": "bytes[]"
            }
          ],
          "internalType": "struct Executor.Proposal[]",
          "name": "proposals",
          "type": "tuple[]"
        },
        {
          "internalType": "bytes[]",
          "name": "accountProof",
          "type": "bytes[]"
        },
        {
          "internalType": "uint256",
          "name": "slot",
          "type": "uint256"
        }
      ],
      "name": "executeProposals",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "uint8",
          "name": "domainID",
          "type": "uint8"
        },
        {
          "internalType": "uint256",
          "name": "depositNonce",
          "type": "uint256"
        }
      ],
      "name": "isProposalExecuted",
      "outputs": [
        {
          "internalType": "bool",
          "name": "",
          "type": "bool"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "uint8",
          "name": "",
          "type": "uint8"
        },
        {
          "internalType": "uint256",
          "name": "",
          "type": "uint256"
        }
      ],
      "name": "usedNonces",
      "outputs": [
        {
          "internalType": "uint256",
          "name": "",
          "type": "uint256"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    }
  ]`
