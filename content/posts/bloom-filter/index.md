---
title: "The Magic of Bloom Filters"
description: "A deep dive into Bloom Filters in Go. We build one from scratch, optimize it with Kirsch-Mitzenmacher, and discover why your choice of hash functions might not matter as much as you think. Or... do they?"
tags:
  - Go
  - Algorithms
  - Data Structures
  - Performance
  - Security
lastUpdated: 2026-03-01
---

# The Magic of Bloom Filters

*A deep dive into Bloom Filters in Go. We build one from scratch, optimize it with Kirsch-Mitzenmacher, and discover why your choice of hash functions might not matter as much as you think. Or... do they?*

<!-- 
Image prompt: A surreal, retro-futuristic pixel art scene of a giant, glowing mechanical sieve or filter floating in space. A chaotic stream of binary data and light particles (representing information) is pouring into the top of the sieve. Only specific, colorful particles are passing through the fine mesh, while the rest bounce off into the void. The color palette should be vibrant neon synthwave (cyan, magenta, deep purple) with a 16-bit aesthetic. 
-->
![Bloom Filter Hero Image](./cover.webp)

## The "Hmmm, Probably?" Data Structure

I recently needed a way to check if a key already existed in a dataset for a project. My first instinct was the classic move: *cache it! Throw a Hash Map at it!* But then I started thinking about scale. If you have millions of keys and you store every single one in a traditional cache, you're going to burn gigabytes of expensive RAM just to hold onto strings you *might* need to query. And what happens when someone queries a key that *doesn't* exist? Your cache misses, you hit the database, search the index, and *then* return a 404. Thousands of times a second.

This isn't a hypothetical problem. In distributed systems, this pattern is called [cache penetration](https://www.tencentcloud.com/techpedia/138162) — where queries for non-existent data bypass the cache and hammer the database directly. It's such a well-known problem that [AWS explicitly recommends Bloom filters](https://aws.amazon.com/es/blogs/database/implement-fast-space-efficient-lookups-using-bloom-filters-in-amazon-elasticache/) as the mitigation in ElastiCache. Attackers don't need to find valid data — they just need to overwhelm your backend with misses. Your database I/O goes through the roof.

We need a way to answer the question *"Does this key exist?"* instantly, in memory, without storing massive strings or allocating gigabytes. This is exactly how [Apache Cassandra skips SSTables](https://cassandra.apache.org/doc/latest/cassandra/managing/operating/bloom_filters.html) without touching the disk, how [Google Bigtable avoids disk lookups](https://research.google/pubs/bigtable-a-distributed-storage-system-for-structured-data/) for non-existent rows, how [Ethereum's `logsBloom` (pdf)](https://ethereum.github.io/yellowpaper/paper.pdf) lets clients scan block headers instead of full block bodies, and how [Mozilla Firefox's CRLite](https://blog.mozilla.org/security/2020/01/09/crlite-part-1-all-web-pki-revocations-compressed/) compresses the entire internet's revoked certificates into a ~10 MB download.

## Enter the Bloom Filter

A [Bloom filter](https://en.wikipedia.org/wiki/Bloom_filter) is a simple, space-efficient probabilistic data structure, originally described by [Burton Howard Bloom in 1970](https://dl.acm.org/doi/10.1145/362686.362692). By "probabilistic", I mean it can't tell you exactly *what* is in a set, but it can answer the question *"Is this item in the set?"* with a very specific guarantee:

- **"No."** (And it is absolutely, 100% right).
- **"Hmmm, Probably?"** (It might be right, or it might be a *false positive*).

There are **no false negatives** in a Bloom filter. If the filter says a key doesn't exist, the key definitively doesn't exist. You just saved a disk read, a network hop, and/or a database query. For Cassandra, this means skipping entire SSTables without touching the disk. For Ethereum, it means scanning compact block headers instead of replaying every transaction.

What I thought would be a Saturday afternoon project turned into three days of staring at benchmark numbers and questioning everything I knew about hashing :sweat_smile:. Let's dive in.

## The VIP Bouncer

Think of a Bloom filter like a wildly efficient bouncer at an exclusive nightclub.

When you walk up to the door and give him your ID (the data), he doesn't check a giant master list of all 10,000 VIP members. Instead, he just has a single sheet of paper with thousands of tiny, numbered checkboxes (the `bits`).

He also has three completely different mathematical formulas in his head (the `Hash Functions`) that he applies to your name:

1. **Formula A** says: your name maps to box **#42**.
2. **Formula B** says: your name maps to box **#105**.
3. **Formula C** says: your name maps to box **#12**.

If you are actually on the VIP list, the bouncer *already checked* all three of those exact boxes earlier in the day when the list was created. 

When you show up, he looks at those three specific boxes on his sheet. If **even one of them is empty**, he knows for an absolute fact you aren't on the list. You are rejected instantly. (This is the definitive `No`).

But, if **all three boxes** are already checked, he lets you in. (This is the `Probably`). 

Why "probably"? Well, the bouncer has been checking boxes for thousands of other people all night! Maybe Pedro caused him to check box #42, Miguel made him check #105, and Carlos caused #12 (yes, this is a latino bar). The boxes are checked, but *not because of you*. You just snuck in based on the overlapping checkmarks of three completely different people. That is a false positive. 

By making the sheet of paper bigger (more bits) or giving him more formulas to calculate (more hashes), we can mathematically control exactly how often someone accidentally sneaks in.

## Building the Filter

Let's build this in Go. 

First, we need the "sheet of checkboxes" — a bit array. The more items we plan to store, the bigger the array needs to be (we'll work out the exact math in a moment). So how do we represent a potentially massive bit array in Go?

The simplest approach would be a boolean slice `[]bool`, but that wastes an entire byte (8 bits) for a single True/False value — eight times more memory than necessary (ouch :weary:). A `[]byte` slice would be more efficient, but for best performance on modern 64-bit architectures, we want to align our memory directly with the CPU word size. Fetching a 64-bit word from memory and masking a single bit is something modern CPUs can do almost instantaneously.

We will use a slice of `atomic.Uint64`. Using atomics gives us concurrent safety for free — multiple goroutines can `Add` and `Contains` simultaneously without a mutex. This is important because in a real service, you'll have hundreds of goroutines hitting the filter concurrently.

We also need to hold onto the size of our bitset (`m`), the number of hash functions we use (`k`), and the actual hashing function we want to invoke.

<<< @/posts/bloom-filter/bloom.go#filter

I love the functional option pattern in Go. It allows us to keep our APIs incredibly clean while remaining open for extension.

<<< @/posts/bloom-filter/bloom.go#hashfunc

Let's create an `Option` type and a specific `WithHashFunc` modifier so users can swap out the hash function if they want to.

<<< @/posts/bloom-filter/bloom.go#options

Notice that we use `hash.Hash64` from the standard library. By programming against an interface rather than a concrete struct, we leave the door open for consumers to inject high-performance hashers like `Murmur3` or `xxHash` later. (Keep this in mind, because it completely bites us in the benchmarks later on :thinking:).

## Creating the Filter

We don't want to enforce arbitrary sizes or force users to guess how many bits they need. We want users to approach us with real-world constraints: *"I have 10,000 items, and I only want a 1% chance of a false positive."* 

The math for this is [well-known](https://en.wikipedia.org/wiki/Bloom_filter#Optimal_number_of_hash_functions). We can calculate the optimal number of bits (`m`) and the optimal number of hash functions (`k`) using natural logarithms. The formulas come directly from [Bloom's original 1970 paper](https://dl.acm.org/doi/10.1145/362686.362692), and the derivations are covered thoroughly in [Mitzenmacher & Upfal's *Probability and Computing* (pdf)](https://www.cs.purdue.edu/homes/spa/courses/pg17/mu-book.pdf) textbook.

<<< @/posts/bloom-filter/bloom.go#math

Why do we need this balance? If your bit array (`m`) is too small, you'll end up checking every single box on the sheet of paper. Once every box is checked, the filter is useless — every query returns `True` (100% false positive rate). And if you use too many hash functions (`k`), you set too many bits per item, filling up your array even faster (it's a lose-lose :weary:).

Armed with that math, our constructors look incredibly clean. 

<<< @/posts/bloom-filter/bloom.go#newfilter

Notice in `NewFilter` how we align the requested bits into our `[]uint64` layer. Because `len(bitset)` operates on 64-bit words, we divide `m` by 64. We intentionally add 63 before dividing to ensure integer arithmetic rounds *up*, guaranteeing we never allocate slightly less space than we need!

## The Kirsch-Mitzenmacher Optimization

A Bloom filter requires `k` *independent* hash functions. If our math says `k` should be 7, do we literally import 7 different hashing algorithms into our codebase? No, that's expensive and impractical (and honestly kind of ridiculous :sweat_smile:).

The standard solution is the [Kirsch-Mitzenmacher optimization (pdf)](https://www.eecs.harvard.edu/~michaelm/postscripts/rsa2008.pdf), published in 2006. It proves mathematically that you only need **two** hash functions to simulate any number of independent hashes. Here's how it works:

You hash the data *once* using a strong 64-bit hasher. You then split that 64-bit result cleanly down the middle into two 32-bit halves (let's call them `h1` and `h2`). You can then simulate an infinite number of totally independent hashes using a simple formula: `hash_i = h1 + i * h2`. That's it. No extra imports, no expensive re-hashing.

Now, where this gets *experiential* — when I first implemented the filter, I tried a naive approach: take the `[]byte` data, append the loop counter (`i`) as a "seed", and hash it `k` times. I thought this would simulate independent hashes well enough. **It did not.** My validation tests expected a 1% false positive rate, and I got 9.8%. Why? Because appending a single sequential byte and re-hashing produces highly correlated bit patterns — the hashes clump together on the same bits, defeating the entire mathematical purpose of the filter.

When I applied Kirsch-Mitzenmacher, the false positive rate dropped from 9.8% straight down to ~1.1%. And because we only call `.Write()` and `.Sum64()` *once* instead of `k` times, the performance skyrocketed. You can see this applied directly in the `Add` and `Contains` methods below.

So we need **one** good 64-bit hash. Which one? If you look at the [xxHash benchmarks](https://xxhash.com/#benchmarks), the throughput hierarchy is pretty clear: XXH3 tops everything at 31.5 GB/s, XXH64 follows at 19.4 GB/s, Murmur3 at 3.9 GB/s, and FNV64 trails at 1.2 GB/s. And it's not just speed — the xxHash benchmark suite scores hash *quality* on a 1-10 scale. XXH3, XXH64, and Murmur3 all score a perfect 10, while FNV64 gets a 5 with the note *"poor avalanche properties"*. All three challengers have Go implementations with hand-tuned assembly for amd64 ([`zeebo/xxh3`](https://github.com/zeebo/xxh3), [`cespare/xxhash`](https://github.com/cespare/xxhash), and [`twmb/murmur3`](https://github.com/twmb/murmur3)), so on paper, the choice should be obvious.

But I defaulted to `hash/fnv` from Go's standard library anyway :sweat_smile:. Zero dependencies, ships with every Go installation, and for a Bloom filter — where you're hashing tiny 16-byte keys, not streaming gigabytes — the raw throughput gap shrinks considerably. We'll put all four head-to-head in the benchmarks later and see if that "poor quality" score actually matters in practice.

## Adding and Checking Data

With Kirsch-Mitzenmacher cleanly handled, the actual `Add` and `Contains` logic becomes elementary bitwise math. 

When we `Add`, we allocate a fresh hasher, hash the data **once**, split the 64-bit result into two 32-bit halves right there, and then iterate `k` times. On each loop, we simulate the next hash via `h1 + i*h2`, modulo it into our `m` boundary, and flip the bit. We use `atomic.Uint64.Or()` to set bits and `atomic.Uint64.Load()` to read them, making the entire filter **safe for concurrent use** without a mutex.

<<< @/posts/bloom-filter/bloom.go#add

We divide the bit index by 64 to find the right word, modulo by 64 to find the bit within that word, and atomically OR (`|`) it to 1 without touching its neighbors. Clean, compact, and lock-free.

Notice something important here: we only ever flip bits **to 1**, never back to 0. That means once a bit is set, it stays set forever. If you're already thinking "wait, does that mean you can't delete items?" — yes, exactly. Multiple items can share the same bit positions, so unsetting a bit to remove one item could corrupt the membership of another. We'll come back to how real systems deal with this constraint shortly.

Checking if data exists (`Contains`) is the exact same logic in reverse.

<<< @/posts/bloom-filter/bloom.go#contains

Instead of setting the bit, we atomically load the word and use Bitwise AND (`&`) to isolate the bit. If the result is `0`, that bit was never set — we short-circuit and return `false` instantly!

## Putting It Through Its Paces

Ok, we have code. Now let's throw some numbers at it. All benchmarks use **16-byte keys** (the size of a UUID — the most common key format you'll see in practice) and a sliding-window over a single random buffer to keep allocation noise out of the measurements. Ryzen 7 5800X, 16 threads, `go test -bench=. -benchmem -benchtime=3s`. Let's go.

### Insert 1M Keys

First question: how long does it take to create a filter for 1 million items at 1% FPR and shove all of them in?

```text
BenchmarkFilter_Insert1M-16    72    48876610 ns/op    9585059 bits    1198136 bytes/filter    7 hash_fns
```

~48.9 ms to insert **1 million UUID-sized keys** — that's roughly **48.9 nanoseconds per key**. The filter itself occupies **1.14 MB** of memory (`1198136 bytes / 1024² ≈ 1.14 MB`), which matches the theoretical formula perfectly: `m ≈ 9.58 million bits`. For comparison, storing 1 million 16-byte keys in a Hash Map would cost you at minimum ~16 MB of raw key data, plus the map overhead (bucket pointers, hash tables) which balloons it to ~50-80 MB depending on load factor. The Bloom filter is **40-70× smaller**.

### Lookup Speed at Scale

Now the question that actually matters in production — with 1M items already packed in, how fast are lookups?

```text
BenchmarkFilter_Contains1M-16    76990525    45.81 ns/op    8 B/op    1 allocs/op
```

**45.81 nanoseconds** per lookup. That's roughly **22 million lookups per second** on a single core. Not bad for a data structure that fits in 1.14 MB :astonished:. The sole 8-byte allocation per call is the FNV hasher interface (we'll come back to this in a minute).

### Concurrent Stress Test

Here's where the `atomic.Uint64` design pays off. I spun up `GOMAXPROCS` goroutines (16 on my machine) doing a 50/50 mix of `Add` and `Contains` on a shared filter with 500K items already in it:

```text
BenchmarkFilter_ConcurrentAddContains-16    459578769    7.401 ns/op    8 B/op    1 allocs/op
```

**7.4 ns/op** under full contention across 16 goroutines. That's over **135 million concurrent operations per second**. No mutexes, no lock contention, no degradation. The `atomic.Uint64` bet paid off big time. This is exactly the behavior you want when hundreds of goroutines are hammering the filter in production.

### Does the Math Actually Hold Up?

All these benchmarks are great, but the whole point of the filter is the false positive rate. Does our 1% target actually work at scale? And here's the fun question — does the *quality* of the hash function matter? I inserted 1M UUID-sized keys, then queried 1M keys that were *never* inserted, using **all four hash functions** on the exact same data:

```text
Filter size (m): 9585059 bits (1.14 MB)
Hash functions (k): 7

=== FPR Validation [FNV] (1M items) ===
  False positives: 9893
  Target FPR: 1.0000%
  Actual FPR: 0.9893%

=== FPR Validation [xxHash] (1M items) ===
  False positives: 10003
  Target FPR: 1.0000%
  Actual FPR: 1.0003%

=== FPR Validation [Murmur3] (1M items) ===
  False positives: 9924
  Target FPR: 1.0000%
  Actual FPR: 0.9924%

=== FPR Validation [XXH3] (1M items) ===
  False positives: 10038
  Target FPR: 1.0000%
  Actual FPR: 1.0038%
```

All four land within ~0.02% of the 1% target. That's basically noise. So the hash "quality" argument? FNV scores 5/10 on the [xxHash benchmark suite](https://xxhash.com/#benchmarks) with the note *"poor avalanche properties"*, while xxHash, Murmur3, and XXH3 all score a perfect 10 — and yet they all produce identical false positive rates. With Kirsch-Mitzenmacher splitting a 64-bit hash into two 32-bit halves, even FNV's simpler distribution is plenty. The math works :tada:.

## The Interface Allocation Plot Twist

Ok so if hash quality doesn't matter... which one is *fastest*? I ran `FNV`, `xxHash`, `Murmur3`, and `XXH3` head-to-head on the same 16-byte UUID keys:

```text
BenchmarkFilter_Add_FNV-16             	 88184716	    39.07 ns/op	     8 B/op	  1 allocs/op
BenchmarkFilter_Add_xxHash-16          	 66102469	    45.68 ns/op	    80 B/op	  1 allocs/op
BenchmarkFilter_Add_Murmur3-16         	 68050894	    52.44 ns/op	    96 B/op	  1 allocs/op
BenchmarkFilter_Add_XXH3-16            	 21895994	   153.5  ns/op	  1280 B/op	  1 allocs/op
```

`FNV` wins by a landslide. And XXH3 — the *fastest hash in the world* by streaming benchmarks — is dead last at nearly **4× slower** than FNV. What happened?!

The `-benchmem` column tells the whole story. Look at `B/op`: `FNV` allocates **8 bytes** per operation. `xxHash`? **80 bytes**. `Murmur3`? **96 bytes**. `XXH3`? A whopping **1280 bytes**. They all show `1 allocs/op`, but the *size* of each allocation is wildly different.

**The Interface Allocation Penalty**: Remember our `HashFunc` type (`HashFunc func() hash.Hash64`)? It returns a standard library interface. Because it returns an interface rather than a concrete struct, the Go compiler can't prove where the struct memory lives or how large it will be. This means a heap allocation *every single time* we call `f.hasher()` inside `Add` or `Contains`. `FNV` creates a tiny 8-byte struct. `xxHash` and `Murmur3` create heavier wrappers. And `XXH3`? Its internal state for the SIMD-optimized algorithm needs **1280 bytes** just to exist — a massive struct that gets allocated and thrown away on every. single. call. The very thing that makes XXH3 blazing fast on large data (wide internal SIMD state) makes it catastrophically expensive to construct for small keys. (So the "flexibility" of our interface design is literally costing us performance... ironic, right? :sweat_smile:)

The `Contains` benchmarks tell the same story:

```text
BenchmarkFilter_Contains_FNV-16        	160837656	    22.20 ns/op	     8 B/op	  1 allocs/op
BenchmarkFilter_Contains_xxHash-16     	123866376	    28.99 ns/op	    80 B/op	  1 allocs/op
BenchmarkFilter_Contains_Murmur3-16    	 91996884	    38.76 ns/op	    96 B/op	  1 allocs/op
BenchmarkFilter_Contains_XXH3-16       	 25076642	   145.9  ns/op	  1280 B/op	  1 allocs/op
```

For our use case — UUID-length keys with Kirsch-Mitzenmacher — `FNV` gives us the best speed, the smallest allocations, and (as we saw above) identical false positive rates. The streaming benchmarks on [xxhash.com](https://xxhash.com/#benchmarks) are measuring a completely different workload. When you're hashing 16 bytes at a time, the cost of *creating* the hasher dominates the cost of *running* it.


## Living With the No-Delete Constraint

We already saw why deletion is off the table — bits are shared across items, so flipping them back to `0` would introduce false negatives. But this constraint also comes with a hidden upside: a Bloom filter stores **no actual data**, just bits. You can't reverse-engineer a bit array back into usernames. Even if an attacker gets a full memory dump, all they have is a bunch of `1`s and `0`s with no way to reconstruct the original data. And remember that cache penetration attack from the intro? With a Bloom filter in front, those millions of fake usernames get rejected *instantly* in memory. For an attacker to even *reach* your database, they'd need to guess a username that produces a false positive — and that only happens 1% of the time.

So what do you do when your data changes? A few strategies, each with real implementations you can study:

1. **Rebuild from scratch.** If your dataset is relatively static, just rebuild the filter periodically. This is exactly what [Apache Cassandra does during compaction](https://cassandra.apache.org/doc/latest/cassandra/managing/operating/bloom_filters.html) — when SSTables are merged, the Bloom filters are rebuilt from the new data. For 1 billion items at a 1% false positive rate, the math gives us `m ≈ 9.58 billion bits` — that's roughly **1.12 GB**. Sounds like a lot, but rebuilding a 1 GB bitset from a database dump is way cheaper than you'd think. And compared to storing those billion strings in a Hash Map? You'd need **tens of gigabytes** depending on average key length. The Bloom filter is an order of magnitude smaller.

2. **Counting Bloom Filters.** Instead of storing a single bit per slot, store a small counter (usually 4 bits). When you add an item, increment the counters. When you delete, decrement. This was formalized by [Fan, Cao, Almeida & Broder (2000)](https://pages.cs.wisc.edu/~jussara/papers/00ton.pdf) and there are production-ready Go implementations like [`tylertreat/BoomFilters`](https://github.com/tylertreat/BoomFilters). The tradeoff? Your filter is now 4× the memory — so that 1 billion item filter jumps from ~1.12 GB to ~4.5 GB. Still way smaller than a full string cache, but it's a real cost. (Nothing is free, right? :sweat_smile:)

3. **Time-based rotation.** Run two filters: a "current" and a "previous" generation. All new writes go to "current". Queries check both. Periodically, toss the old one and promote "current" to "previous". This is Akamai's ["one-hit-wonder" filter (pdf)](https://www.eecs.harvard.edu/~michaelm/postscripts/im2005b.pdf) in action — a CDN Bloom filter tracks which URLs have been requested at least once, and only on the *second* request does the CDN actually write to disk cache. Since ~75% of web objects are requested exactly once, this avoids polluting the cache with ephemeral junk. [Apache Nutch](https://nutch.apache.org/) uses a similar rotation for web crawl deduplication — you don't care about URLs you saw 3 days ago, you just don't want to re-crawl the same page twice in the same pass.

## Recapitulation

The standard library `FNV` hasher, a clean `uint64` bitset, and Kirsch-Mitzenmacher. That's it. 1M UUID keys inserted in ~49ms. Lookups in under 46ns. ~118 million concurrent adds/second and ~135 million mixed ops/second across 16 goroutines. All in 1.14 MB of memory. And all three hash functions we tested — FNV, xxHash, Murmur3 — hit within 0.02% of the target FPR. Hash quality is a non-factor.

A simple, standard library hash function, allocated once, mapped cleanly across a `uint64` slice, evaluates millions of items in fractions of a millisecond. Sometimes simple wins.

Fun footnote: our 1.14 MB filter fits comfortably in L3 cache, and that's *partly* why these numbers look so good — each lookup is just a few random bit checks against cached memory. What happens when a filter *doesn't* fit? [Cloudflare found out](https://blog.cloudflare.com/when-bloom-filters-dont-bloom/) when they pushed a 128 MB filter with 8 hash functions against billions of lines and each probe turned into a full RAM fetch. But that's Cloudflare-scale territory — most practical filters (even our 1-billion-item projection at 1.12 GB) will either fit in cache or sit behind an I/O boundary where the filter's memory footprint is still orders of magnitude cheaper than the alternative. For the vast majority of use cases, Bloom filters just work.

The next time you're sketching an architecture and instinctively reach for a caching layer to fix database load, ask yourself: *Do I need to cache the data itself, or do I just need to know if the data exists?* 

Sometimes, all you need is a "Hmmm, Probably?".

Arrivederci! :wave:

## Further Reading

- [Space/Time Trade-offs in Hash Coding with Allowable Errors](https://dl.acm.org/doi/10.1145/362686.362692) — Burton H. Bloom's original 1970 paper that started it all.
- [Less Hashing, Same Performance: Building a Better Bloom Filter (pdf)](https://www.eecs.harvard.edu/~michaelm/postscripts/rsa2008.pdf) — Kirsch & Mitzenmacher (2006), the proof that two hash functions are enough.
- [Bloom filter — Wikipedia](https://en.wikipedia.org/wiki/Bloom_filter) — Excellent overview with derivations for optimal `m` and `k`.
- [Summary Cache: A Scalable Wide-Area Web Cache Sharing Protocol (pdf)](https://pages.cs.wisc.edu/~jussara/papers/00ton.pdf) — Fan, Cao, Almeida & Broder (2000), the Counting Bloom Filter paper.
- [Network Applications of Bloom Filters: A Survey (pdf)](https://www.eecs.harvard.edu/~michaelm/postscripts/im2005b.pdf) — Broder & Mitzenmacher (2004), the definitive survey covering Akamai, distributed caching, and more.
- [Bloom Filters by Example](https://llimllib.github.io/bloomfilter-tutorial/) — Interactive visualization that lets you see bits being set in real time.
- [Bigtable: A Distributed Storage System for Structured Data](https://research.google/pubs/bigtable-a-distributed-storage-system-for-structured-data/) — Google's 2006 paper; Section 7 details their Bloom filter usage.
- [Apache Cassandra — Bloom Filters](https://cassandra.apache.org/doc/latest/cassandra/managing/operating/bloom_filters.html) — How Cassandra uses Bloom filters to skip SSTables.
- [Introducing CRLite](https://blog.mozilla.org/security/2020/01/09/crlite-part-1-all-web-pki-revocations-compressed/) — Mozilla's cascading Bloom filters for certificate revocation.
- [Ethereum Yellow Paper (pdf)](https://ethereum.github.io/yellowpaper/paper.pdf) — The `logsBloom` field specification for event log filtering.
- [BIP 37: Connection Bloom Filtering](https://github.com/bitcoin/bips/blob/master/bip-0037.mediawiki) — Bitcoin's (now deprecated) use of Bloom filters for SPV wallets.
