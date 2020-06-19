# Elrond Auto Delegate

Copy `config.yml` to `mainnet.yml` and edit configuration.

## Approve Genesis contract to use your ERD tokens

```
go run ./cmd/approve -c mainnet.yml
```

## Delegate for the first time your ERD tokens

```
go run ./cmd/delegate -c mainnet.yml
```

## Increase amount of delegated ERD tokens

```
go run ./cmd/increase -c mainnet.yml
```

## Regenerate Smart Contract Go code

```
solc --abi genesis.sol
solc --bin genesis.sol
abigen --bin genesis_sol_GenesisSC.bin --abi genesis_sol_GenesisSC.abi --pkg genesis --out genesis.go
```

```
solc --abi token.sol
solc --bin token.sol
abigen --bin token_sol_ERDToken.bin --abi token_sol_ERDToken.abi --pkg token --out token.go
```