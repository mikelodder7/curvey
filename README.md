# curvey
Implementations of constant time elliptic curves in go. The following elliptic curves are provided

- BLS12-381
- Ed25519
- Secp256k1
- NistP256
- Pallas

These curves all implement a common interface and as such can be used in a curve agnostic manner.

This work is derived from Coinbase's kryptology library.


## Security and Compliance

All the code is strict, both in terms of timing-based side-channels (everything is constant-time, except if explicity state otherwise) and in compliance to relavant standards. There is no attempt at "zeroing memory" anywhere in the code. Golang just doesn't really support this easily and it ruins readability. In general, such memory cleansing is a fool's errand anyway. 

**Warning**: Although all of the code aims at being representative of optimized production ready-code, some bugs might still exist, no matter how careful I am. Any assertion of suitability to any purpose is explicitly denied. The primary purpose of this library is to enable various elliptic curves that implement a common interface, and easy-to-use API.

## License

Licensed under Apache License, Version 2.0, [LICENSE](./LICENSE) or http://www.apache.org/licenses/LICENSE-2.0

## Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in the work by you, as defined in the Apache-2.0 license, shall be licensed as above, without any terms or conditions.
