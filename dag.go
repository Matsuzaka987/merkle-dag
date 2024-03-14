package merkledag

import "hash"

// Add adds a node to the store and returns its hash.
func Add(store KVStore, node Node, h hash.Hash) []byte {
    // Compute the hash for a single node.
    computeHash := func(n Node) []byte {
        h.Reset()
        h.Write(n.Bytes())
        return h.Sum(nil)
    }

    // Pair two hashes and return a new node representing their combination.
    pairHashes := func(hash1, hash2 []byte) Node {
        h.Reset()
        h.Write(hash1)
        h.Write(hash2)
        pairedHash := h.Sum(nil)
        store.Put(pairedHash, append(hash1, hash2...)) // Store the combined hash.
        return newNode(pairedHash)
    }

    // newNode creates a new node from a hash.
    newNode := func(hash []byte) Node {
        return &pair{
            size:  uint64(len(hash)),
            bytes: hash,
        }
    }

    // Define the pair struct to implement the Node interface.
    type pair struct {
        size  uint64
        bytes []byte
    }

    // Implement the Node interface methods for pair.
    func (p *pair) Size() uint64 { return p.size }
    func (p *pair) Type() int    { return FILE } // Assuming all pairs are treated as files.
    func (p *pair) Bytes() []byte { return p.bytes }

    // Process the node based on its type.
    switch node.Type() {
    case FILE:
        hash := computeHash(node)
        store.Put(hash, node.Bytes()) // Store the file hash.
        return hash
    case DIR:
        dir := node.(Dir)
        var pairs []Node

        // Process each node in the directory.
        for it := dir.It(); it.Next(); {
            pairs = append(pairs, computeHash(it.Node()))
        }

        // Pair the hashes until one root hash remains.
        for len(pairs) > 1 {
            var newPairs []Node
            for i := 0; i < len(pairs); i += 2 {
                if i+1 < len(pairs) {
                    newPairs = append(newPairs, pairHashes(pairs[i].Bytes(), pairs[i+1].Bytes()))
                } else {
                    newPairs = append(newPairs, pairs[i]) // Handle the case of an odd number of pairs.
                }
            }
            pairs = newPairs
        }

        return pairs[0].Bytes() // Return the root hash.
    default:
        panic("unsupported node type")
    }
}