# Hashing Algorithms Performance Benchmark

This sub-project benchmarks various hashing algorithms implemented in Go to determine the best fit for different use cases (Password storage vs. API Key verification).

---

## Hashing Algorithms Comparison Table

| Algorithm       | Category      | Speed          | Resource Intensity  | Security Level | Best Use Case                                                     |
| :-------------- | :------------ | :------------- | :------------------ | :------------- | :---------------------------------------------------------------- |
| **Argon2 (id)** | Password Hash | **Very Slow**  | Very High (Tunable) | **Highest**    | Modern user password storage (Current industry standard).         |
| **Bcrypt**      | Password Hash | **Slow**       | Medium (CPU)        | **High**       | Standard web application passwords (Reliable & widely supported). |
| **Scrypt**      | Password Hash | **Slow**       | High (RAM)          | **High**       | Cryptocurrency wallets, sensitive file backups.                   |
| **SHA-256**     | Cryptographic | Fast           | Low                 | **Good**       | **API Keys**, Blockchain, Digital Signatures, SSL Certificates.   |
| **SHA-3**       | Cryptographic | Fast           | Low                 | **Very High**  | High-security financial or government systems (NIST standard).    |
| **BLAKE3**      | Cryptographic | **Ultra Fast** | Low                 | **Good**       | Large file hashing, High-throughput API Gateways.                 |
| **MD5 / SHA-1** | Legacy        | Very Fast      | Very Low            | **Broken**     | Legacy checksums only. **Do not use for security.**               |
| **XxHash**      | Non-Crypto    | **Insane**     | Very Low            | **None**       | In-memory Cache keys (Redis), Hash Maps in code.                  |
| **MurmurHash3** | Non-Crypto    | Extreme        | Very Low            | **None**       | Big Data, Bloom Filters, Data Sharding/Distribution.              |
| **FarmHash**    | Non-Crypto    | Extreme        | Very Low            | **None**       | Google-scale hash tables, optimized for 64-bit CPUs.              |

## 💻 Environment

- **OS**: Windows (amd64)
- **CPU**: Intel(R) Core(TM) i5-7400 CPU @ 3.00GHz
- **Go Version**: 1.24+ (using `b.Loop()`)

## 📊 Benchmark Results

Executed with: `go test -bench=. -benchmem -benchtime=5s`

| Algorithm    | Category   | Iterations | Time (ns/op)  | Memory (B/op) | Allocs (op) |
| :----------- | :--------- | :--------: | :-----------: | :-----------: | :---------: |
| **Argon2id** | Slow Hash  |     66     | 89,716,462 ns | 67,117,179 B  |     72      |
| **Bcrypt**   | Slow Hash  |     84     | 68,601,094 ns |    5,232 B    |     10      |
| **Scrypt**   | Slow Hash  |    128     | 45,975,688 ns | 16,781,989 B  |     26      |
| **SHA-256**  | Balanced   | 8,489,791  |   689.4 ns    |     224 B     |      4      |
| **SHA-3**    | Balanced   | 9,479,018  |   637.6 ns    |     128 B     |      2      |
| **BLAKE3**   | Balanced   | 2,306,468  |  2,607.0 ns   |   11,040 B    |      4      |
| **XxHash**   | Super Fast | 39,586,348 |   158.3 ns    |     32 B      |      2      |
| **FarmHash** | Super Fast | 40,760,869 |   152.3 ns    |     32 B      |      2      |
| **MurMur3**  | Super Fast | 29,889,914 |   202.4 ns    |     64 B      |      3      |

---

## 🔍 Key Findings

### 1. The "Argon2" Trap for API Auth
- **Performance**: Argon2 takes **~90ms** per operation. In a high-concurrency API environment, a single CPU core could only handle ~11 requests/sec.
- **Memory**: It consumes **67MB** of RAM per request. 100 concurrent auth requests would spike memory usage to **6.7GB**, likely triggering an OOM (Out of Memory) crash.
- **Verdict**: Never use Argon2/Bcrypt/Scrypt for per-request API Key verification.

### 2. Why SHA-256/SHA-3 is the Standard
- **Efficiency**: SHA-256 is **~130,000x faster** than Argon2.
- **Scalability**: It uses negligible memory (224 bytes), allowing the server to handle millions of verifications without breaking a sweat.
- **Verdict**: Best suited for API Keys and Digital Signatures.

### 3. Non-Cryptographic Speed
- **Speed**: XxHash and FarmHash are the winners for pure speed (~150ns).
- **Verdict**: Use these only for internal data structures (Hash Maps, Bloom Filters) or In-memory caching where security against collision attacks is not a priority.

---

## 🚀 How to Run
To reproduce these results on your machine, run:
```bash
go test -bench=. -benchmem -benchtime=5s
```

---

## Usage Cheat Sheet

### 1. Storing User Passwords

- **Primary Choice:** `Argon2id` (Best protection against GPU/ASIC cracking).
- **Alternative:** `Bcrypt` (Simple to implement, natively supported by most frameworks).

### 2. Generating & Verifying API Keys

- **Primary Choice:** `SHA-256` (Perfect balance of speed and security for every-request verification).
- **Alternative:** `BLAKE3` (Use this if you process billions of requests and need to minimize CPU overhead).

### 3. Internal Data Structures (In-memory)

- **Selection:** `XxHash` or `MurmurHash3`.
- **Why:** Speed is the only priority. Security is not required because the data stays within your internal application memory.

### 4. Verifying File Integrity

- **Selection:** `SHA-256` (Secure) or `BLAKE3` (Extremely fast for multi-gigabyte files).

### 5. Data Sharding & Distribution

- **Selection:** `MurmurHash3` or `FarmHash`.
- **Why:** They provide excellent "avalanche effect" (uniform distribution), ensuring data is spread evenly across servers to avoid "hot spots."

---

## ⚠️ Developer Rules to Remember

1.  **Never Roll Your Own Hash:** Always use standard, peer-reviewed libraries.
2.  **Slow for Humans, Fast for Machines:**
    - **User passwords** MUST use **Slow** algorithms (to prevent brute-force).
    - **API Keys** (called by machines) MUST use **Fast** algorithms (to prevent system bottlenecks).
3.  **Salting:** Argon2 and Bcrypt handle salting automatically. For other cryptographic hashes used for passwords, always use a unique, random salt.
4.  **No Security in Non-Crypto Hashes:** Never use XxHash, MurmurHash, or CityHash for anything sensitive; they are vulnerable to collision attacks.
