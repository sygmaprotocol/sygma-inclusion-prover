package abi

const StateRootStorageABI = `[
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "uint8",
				"name": "sourceDomainID",
				"type": "uint8"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "slot",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "bytes32",
				"name": "stateRoot",
				"type": "bytes32"
			}
		],
		"name": "StateRootSubmitted",
		"type": "event"
	}
]`
