# ZTYP

Another approach to SSZ, focused on typing around merkle-tree-representations of state.
For a super-fast but non-tree approach, try [ZSSZ](https://github.com/protolambda/zssz).

Work in progress. No serialization/deserialization for types yet.

## Design

Construction:
- Ztyp describes types as interfaces to a binary merkle tree: views over a backing tree.
- Every node in the backing tree is referred to as a `Node`, build with:
    - `Commit`: combines two subtrees
    - `Root`: a bytes32. Can be a summary root of an unexpanded subtree, or a just a leaf node.
    - The backing is immutable: on modification the tree is forked, with both forks sharing the common parts
    - The full tree is cached automatically! Since child nodes never change their identity,
       the hash-tree-root result is cached in every merkle-tree node. *Without the need to even know the type*
- a `TypeDef` and `BasicTypeDef` are implemented to describe the raw type information.
    - Create default trees from the type-definition. No views required.
    - Convert an untyped backing into a typed view by attaching the type definition.
    - Convert a `Root` and sub-index into a typed sub-view by attaching a basic type definition.
- a `View` is used to interact with a subtree. A general view only tracks its backing.
    - Typed views allow you to interact with a backing in typed ways:
        - Basic types: `Uint64Type`, `Uint32Type`, `Uint16Type`, `Uint8Type`, `BoolType`
        - Composite types: `Container`, `List`, `Vector`
        - Basic composite types (to enable packing of consecutive elements): `BasicList`, `BasicVector`
        - Bitfields: `Bitvector`, `Bitlist`
        - Optimized small byte vectors: `SmallByteVecMeta`: to derive any `BytesN` (`N <= 32`) from.
        - `Root` itself is also a valid view.
    - Semi-typed views are useful to build your own types: `SubtreeView`
- `ReadProp`/`WriteProp` functions can be used to describe reusable `Getter/Setter -> View -> *my-type*` pipelines.
    - Take a typed view that has getters/setters, create closure to fix it to a property index
    - Type the property function and add a receiver func to return the typed view instead.
- The backing tree can be partial, and summarised/expanded dynamically. The typedefinition will safely handle a tree,
   and return an error when the expected data for an operation is inconsistent or missing.
- The hash-function for hash-tree-root is:
    - passed by reference, to reuse a single state during hashing.
    - pluggable. Just define a `H(a [32]byte, b [32]byte) [32]byte` and plug it into `Hash` and `InitZeroHashes`.


## Contact

Dev: [@protolambda on Twitter](https://twitter.com/protolambda)


## License

MIT, see license file.

