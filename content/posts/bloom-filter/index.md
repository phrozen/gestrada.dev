---
title: "The Magic of Bloom Filters"
description: "A deep dive into Bloom Filters in Go. We build one from scratch, optimize it with Kirsch-Mitzenmacher, and discover why your choice of hash functions might not matter as much as you think."
tags:
  - Go
  - Algorithms
  - Data Structures
  - Performance
  - Security
lastUpdated: 2026-02-28
---

<!-- 
Image prompt: A surreal, retro-futuristic pixel art scene of a giant, glowing mechanical sieve or filter floating in space. A chaotic stream of binary data and light particles (representing information) is pouring into the top of the sieve. Only specific, colorful particles are passing through the fine mesh, while the rest bounce off into the void. The color palette should be vibrant neon synthwave (cyan, magenta, deep purple) with a 16-bit aesthetic. 
-->
![Bloom Filter Hero Image](./cover.webp)

## The "Hmmm, Probably?" Data Structure

I recently needed a way to check if a key already existed in a dataset for a side project. My first instinct was the classic move: *cache it! Throw a Hash Map at it!* But then I started thinking about scale. If you have a billion registered users and you store every single username in a traditional cache, you are going to consume gigabytes of expensive RAM just to hold onto strings you *might* need to query. And what happens when someone queries a username that *doesn't* exist? Your cache misses, you hit the database, search the index, and *then* return a 404. Thousands of times a second. Your system melts from the I/O overhead of *checking things that aren't there*.

We need a way to answer this question *instantly*, in memory, without storing massive strings or allocating gigabytes.

## Enter the Bloom Filter

A Bloom filter is a simple, space-efficient probabilistic data structure. By "probabilistic", I mean it can't tell you exactly *what* is in a set, but it can answer the question *"Is this item in the set?"* with a very specific guarantee:

- **"No."** (And it is absolutely, 100% right).
- **"Hmmm, Probably?"** (It might be right, or it might be a *false positive*).

There are **no false negatives** in a Bloom filter. If it says a username doesn't exist, you don't even need to check your database. You just saved a disk read and a network hop! You can immediately tell the user to proceed.

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

Why "probably"? Well, the bouncer has been checking boxes for thousands of other people all night! Maybe Alice caused him to check box #42, Bob made him check #105, and Charlie caused #12. The boxes are checked, but *not because of you*. You just snuck in based on the overlapping checkmarks of three completely different people. That is a false positive. 

By making the sheet of paper bigger (more bits) or giving him more formulas to calculate (more hashes), we can mathematically control exactly how often someone accidentally sneaks in.

## Building the Filter

Let's build this in Go. 

First, we need the "sheet of checkboxes". The simplest way to represent a bit array in Go might seem to be a boolean slice `[]bool`, but that wastes an entire byte (8 bits) for a single True/False value! (eight times more memory than we need... ouch :weary:). We could use a byte slice `[]byte`, but for best performance on modern 64-bit architectures, we want to align our memory directly with the CPU word size. 

We will use a slice of `atomic.Uint64`. Fetching a 64-bit word from memory and masking a single bit is something modern CPUs can do almost instantaneously, and using atomics means we get concurrent safety for free — multiple goroutines can `Add` and `Contains` simultaneously without a mutex. 

We also need to hold onto the size of our bitset (`m`), the number of hash functions we use (`k`), and the actual hashing function we want to invoke.

<<< @/posts/bloom-filter/bloom.go#filter

I love the functional option pattern in Go. It allows us to keep our APIs incredibly clean while remaining open for extension.

<<< @/posts/bloom-filter/bloom.go#hashfunc

Let's create an `Option` type and a specific `WithHashFunc` modifier so users can swap out the hash function if they want to.

<<< @/posts/bloom-filter/bloom.go#options

Notice that we use `hash.Hash64` from the standard library. By programming against an interface rather than a concrete struct, we leave the door open for consumers to inject high-performance hashers like `Murmur3` or `xxHash` later. (Keep this in mind, because it completely bites us in the benchmarks later on :thinking:).

## Creating the Filter

We don't want to enforce arbitrary sizes or force users to guess how many bits they need. We want users to approach us with real-world constraints: *"I have 10,000 items, and I only want a 1% chance of a false positive."* 

Fortunately, the math for this is well-known. We can calculate the optimal number of bits (`m`) and the optimal number of hash functions (`k`) perfectly using natural logarithms.

<<< @/posts/bloom-filter/bloom.go#math

Why do we need this balance? If your bit array (`m`) is too small, you'll end up checking every single box on the sheet of paper. Once every box is checked, the filter is useless — every query returns `True` (100% false positive rate). And if you use too many hash functions (`k`), you set too many bits per item, filling up your array even faster (it's a lose-lose :weary:).

Finding the mathematical sweet spot is critical. Armed with that math, our constructors look incredibly clean. 

<<< @/posts/bloom-filter/bloom.go#newfilter

Notice in `NewFilter` how we align the requested bits into our `[]uint64` layer. Because `len(bitset)` operates on 64-bit words, we divide `m` by 64. We intentionally add 63 before dividing to ensure integer arithmetic rounds *up*, guaranteeing we never allocate slightly less space than we need!

## The Hashing Optimization Trick

Here is where my initial implementation completely fell flat on its face. I want to share this because it was such a big "Aha!" moment for me.

A Bloom filter requires `k` *independent* hash functions. If our math says `k` should be 7, do we literally import 7 different hashing algorithms into our codebase? No, that's expensive and impractical (and honestly kind of ridiculous :sweat_smile:).

At first, I tried a simple loop: take the `[]byte` data, append the loop counter (`i`) as a "seed", and hash it 7 times. I thought this would perfectly simulate 7 independent hashes. **Do not do this.**

When I ran my validation test suite (expecting a 1% false positive rate), my tests failed spectacularly with a 9.8% collision rate! Why? Because appending a single sequential byte to a payload and re-hashing it results in highly correlated bit patterns. The hashes weren't independent at all; they were clumping together on the exact same bits, defeating the entire mathematical purpose of the filter.

I did some digging and discovered the legendary **Kirsch-Mitzenmacher optimization**. 

This brilliant trick (published in 2006) proves you only need to hash the data *once* using a strong 64-bit hasher. You then split that 64-bit result cleanly down the middle into two 32-bit halves (let's call them `h1` and `h2`). 

You can then mathematically simulate an infinite number of totally independent hashes using a simple formula: `hash_i = h1 + i * h2`. That's it. No extra imports, no expensive re-hashing.

It's magical. My false positive rate instantly plummeted from 9.8% down to exactly 1.1%. And because we only call `.Write()` and `.Sum64()` *once* instead of `k` times, the performance skyrocketed. You can see this applied directly in the `Add` and `Contains` methods below.

## Adding and Checking Data

With the Kirsch-Mitzenmacher hash generation cleanly handled, the actual `Add` and `Contains` logic becomes elementary bitwise math. 

When we `Add`, we allocate a fresh hasher, hash the data **once**, split the 64-bit result into two 32-bit halves right there, and then iterate `k` times. On each loop, we simulate the next hash via `h1 + i*h2`, modulo it into our `m` boundary, and flip the bit. We use `atomic.Uint64.Or()` to set bits and `atomic.Uint64.Load()` to read them, making the entire filter **safe for concurrent use** without a mutex.

<<< @/posts/bloom-filter/bloom.go#add

We divide the bit index by 64 to find the right word, modulo by 64 to find the bit within that word, and atomically OR (`|`) it to 1 without touching its neighbors. Clean, compact, and lock-free.

Checking if data exists (`Contains`) is the exact same logic in reverse.

<<< @/posts/bloom-filter/bloom.go#contains

Instead of setting the bit, we atomically load the word and use Bitwise AND (`&`) to isolate the bit. If the result is `0`, that bit was never set — we short-circuit and return `false` instantly!

## The Interface Allocation Plot Twist

When I finished the code, I was pretty proud of it. But I wanted more. I decided to benchmark Go's internal `fnv-1a` hashing algorithm against the heavy hitters of the hashing world: `xxHash` and `Murmur3`. 

I imported them, hooked them into the `WithHashFunc` option, and ran the benchmarks over millions of iterations.

I fully expected `xxHash` to crush `FNV`. I was ready to make `xxHash` the default.

Here is what `go test -bench=. -benchmem` actually spit out:

```text
BenchmarkFilter_Add_FNV-16             	41334127	        30.84 ns/op	   8 B/op	  1 allocs/op
BenchmarkFilter_Add_xxHash-16          	31900555	        40.83 ns/op	  80 B/op	  1 allocs/op
BenchmarkFilter_Add_Murmur3-16         	20587674	        56.11 ns/op	  96 B/op	  1 allocs/op
```

Wait. `FNV` won? By a *landslide*?! :thinking:

How is that even possible? The `-benchmem` output makes it obvious. Look at the `B/op` column: `FNV` allocates **8 bytes** per operation. `xxHash` allocates **80 bytes**. `Murmur3` allocates **96 bytes**. They all show `1 allocs/op` — a single allocation per call — but the *size* of that allocation is wildly different. The answer comes down to two things:

1. **The Interface Allocation Penalty**: Remember our `HashFunc` type (`HashFunc func() hash.Hash64`)? It returns a standard library interface. Because it returns an interface rather than a concrete struct, the Go compiler can't prove where the struct memory lives or how large it will be. This means an expensive heap allocation *every single time* we call `f.hasher()` inside `Add` or `Contains`. `FNV` creates a tiny struct wrapper. `Murmur3` and `xxHash` create much heavier wrappers just to satisfy the interface. (So the "flexibility" of our interface design is literally costing us performance... ironic, right? :sweat_smile:)
2. **Setup Overhead**: We are hashing tiny, 8-byte slices in these benchmarks. `xxHash` and `Murmur3` are fast at hashing large files because they pipeline big blocks of memory through CPU registers. But for a tiny 8-byte payload? The setup and teardown take longer than the actual hashing! `FNV` is basically just a tiny `for` loop (`hash ^= byte; hash *= prime`). It finishes the race before `xxHash` even finishes tying its shoes.

Now, to be fair — `xxHash` and `Murmur3` aren't just about raw speed. They are designed to produce **high-quality, low-collision hashes** with excellent distribution properties. For a Bloom filter operating on millions of unique keys, hash distribution matters a lot — poor distribution means more bit collisions, which means higher false positive rates. If your payloads are larger (URLs, file checksums, serialized objects), the quality advantage of `xxHash` or `Murmur3` might actually outweigh the allocation overhead and make them the better choice. But for tiny keys like the ones in my benchmark? `FNV` is more than good enough, and the speed difference isn't even close.

## The No-Delete Problem

There is one massive gotcha with Bloom filters that I didn't fully appreciate until I tried to use mine in production: **you cannot remove items**. 

Think about it. When we `Add` an item, we flip `k` bits to `1`. But those same bits might have been flipped by *other* items too (that's exactly what causes false positives). If we try to flip them back to `0` to "delete" an item, we might accidentally erase bits that belong to completely different entries. We'd be introducing **false negatives**, which breaks the one guarantee a Bloom filter gives us.

So what do you do when your data changes? A few strategies:

1. **Rebuild from scratch.** If your dataset is relatively static (like a list of known malicious URLs, or a dictionary), just rebuild the filter periodically. For 1 billion items at a 1% false positive rate, the math gives us `m ≈ 9.58 billion bits` — that's roughly **1.12 GB**. Sounds like a lot, but rebuilding a 1 GB bitset from a database dump is way cheaper than you'd think. And compared to storing those billion strings in a Hash Map? You'd need **tens of gigabytes** depending on average key length. The Bloom filter is an order of magnitude smaller.

2. **Counting Bloom Filters.** Instead of storing a single bit per slot, store a small counter (usually 4 bits). When you add an item, increment the counters. When you delete, decrement. This lets you "undo" additions safely. The tradeoff? Your filter is now 4× the memory — so that 1 billion item filter jumps from ~1.12 GB to ~4.5 GB. Still way smaller than a full string cache, but it's a real cost. (Nothing is free, right? :sweat_smile:)

3. **Time-based rotation.** Run two filters: a "current" and a "previous" generation. All new writes go to "current". Queries check both. Periodically, toss the old one and promote "current" to "previous". This is how a lot of web crawlers handle URL deduplication — you don't care about URLs you saw 3 days ago, you just don't want to re-crawl the same page twice in the same pass.

## The Security Angle

Here's something that doesn't get talked about enough: Bloom filters are actually a **security feature**, not just a performance optimization.

Think about what a traditional cache stores. If you're caching usernames for a "does this username exist?" check, your Redis instance now contains a full copy of every username in your system. That's personally identifiable information (PII) sitting in memory, queryable, extractable. If someone compromises your cache layer, they just got your entire user list for free.

A Bloom filter stores **none of that**. It's just bits. You can't reverse-engineer a bit array back into usernames. Even if an attacker gets a full memory dump of your filter, all they have is a bunch of `1`s and `0`s with no way to reconstruct the original data. (That's really reassuring actually :relieved:)

And then there's the attack vector angle. Without a Bloom filter, an attacker can hammer your API with millions of *non-existent* usernames. Every single one misses the cache and hits your database. This is a classic **cache-negative attack** — the attacker isn't trying to find valid data, they're trying to overwhelm your backend with misses. Your database I/O goes through the roof.

With a Bloom filter sitting in front? Those millions of fake usernames get rejected *instantly* in memory. The filter says "No" definitively, and the request never touches your database. Let's put some numbers on it: at 1 billion items with 1% false positive rate, we need `m ≈ 9.58 billion bits` — that's roughly **1.12 GB** using our `uint64` bit-packed array. Compare that to a Hash Map storing a billion username strings — you'd easily need **10-30 GB** of RAM depending on string lengths. The Bloom filter gives you the same "does it exist?" answer at a fraction of the memory, and the attacker can't extract any data from it. For an attacker to even *reach* your database, they'd need to guess a username that produces a false positive — and that only happens 1% of the time. The other 99% of their attack traffic gets absorbed by a ~1 GB memory structure that contains zero actual user data.

## Bloom Filters in the Wild

Bloom filters aren't just an academic curiosity — they're everywhere in production systems, quietly saving billions of unnecessary I/O operations:

- **Apache Cassandra** uses Bloom filters before checking SSTables on disk. When a read request comes in, Cassandra checks the Bloom filter for each SSTable first. If the filter says "No", it skips that entire SSTable without ever touching the disk. Considering a single node might have hundreds of SSTables, this saves an enormous amount of random disk I/O on every single read.
- **Google Chrome's Safe Browsing** uses a local Bloom filter to check if a URL is potentially malicious. Instead of sending every single URL you visit to Google's servers (which would be a privacy nightmare), Chrome checks a local Bloom filter first. Only if the filter says "Probably" does it make a network request to verify. Most of your browsing never leaves your machine.
- **CDNs like Akamai** use Bloom filters to decide whether to cache content. They don't cache an object on the first request — instead, a Bloom filter tracks which URLs have been requested at least once. Only on the *second* request (when the filter confirms it's seen this URL before) does the CDN actually cache the content. This prevents "one-hit wonders" from polluting the cache. (That's actually really clever :astonished:)

## Recapitulation

I honestly didn't expect `FNV` to win. That benchmark humbled me. I had imported `xxHash` fully expecting it to be the obvious default, and the standard library function I almost ignored turned out to be the fastest option for my use case. It's a good reminder that reaching for the "fastest" external library doesn't help much when your specific data pipeline (tiny payloads) and your structural choices (interfaces causing heap allocations) work against it.

A simple, standard library hash function, allocated once, mapped cleanly across a `uint64` slice, evaluates millions of items in fractions of a millisecond. Sometimes simple wins.

Bloom filters are a really fun data structure. The next time you're sketching an architecture and instinctively reach for a caching layer to fix database load, ask yourself: *Do I need to cache the data itself, or do I just need to know if the data exists?* 

Sometimes, all you need is a "Hmmm, Probably?".

Arrivederci! :wave:

<script setup>
import IconBrandGithub from '~icons/tabler/brand-github'
</script>

<div class="source-link">
  <a href="https://github.com/phrozen/gestrada.dev/tree/main/content/posts/bloom-filter/bloom.go" target="_blank" rel="noopener noreferrer">
    <IconBrandGithub /> View source on GitHub
  </a>
</div>

<style>
.source-link {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--vp-c-divider);
}
.source-link a {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--vp-c-text-2);
  text-decoration: none;
  font-size: 0.9rem;
  transition: color 0.2s;
}
.source-link a:hover {
  color: var(--vp-c-brand-1);
}
</style>
