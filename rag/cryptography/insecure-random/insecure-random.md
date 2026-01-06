# Insecure Random Number Generation

**Category**: cryptography/insecure-random
**Description**: Detection of non-cryptographic RNG used for security purposes
**CWE**: CWE-330 (Use of Insufficiently Random Values), CWE-338 (Use of Cryptographically Weak PRNG)

---

## Import Detection

### Python
**Pattern**: `^import random$`
- Non-cryptographic random module (Python's random uses Mersenne Twister)
- Example: `import random`

**Pattern**: `from random import`
- Importing from non-secure random module
- Example: `from random import randint, choice`

**Pattern**: `random\.randint\(`
- Predictable random integer
- Example: `random.randint(0, 100)`

**Pattern**: `random\.randrange\(`
- Predictable random range
- Example: `random.randrange(0, 100)`

**Pattern**: `random\.choice\(`
- Predictable random selection
- Example: `token = ''.join(random.choice(charset) for _ in range(32))`

**Pattern**: `random\.choices\(`
- Predictable random choices
- Example: `random.choices(charset, k=32)`

**Pattern**: `random\.random\(\)`
- Predictable random float [0.0, 1.0)
- Example: `random.random()`

**Pattern**: `random\.uniform\(`
- Predictable random float in range
- Example: `random.uniform(0, 100)`

**Pattern**: `random\.shuffle\(`
- Predictable shuffle
- Example: `random.shuffle(deck)`

**Pattern**: `random\.sample\(`
- Predictable sample selection
- Example: `random.sample(population, k=5)`

**Pattern**: `random\.getrandbits\(`
- Predictable random bits
- Example: `random.getrandbits(128)`

### Javascript
**Pattern**: `Math\.random\(\)`
- JavaScript's predictable Math.random (uses xorshift128+)
- Example: `Math.random()`

**Pattern**: `Math\.floor\(Math\.random\(\)`
- Common pattern for random integers
- Example: `Math.floor(Math.random() * 100)`

**Pattern**: `Math\.round\(Math\.random\(\)`
- Random with rounding
- Example: `Math.round(Math.random() * 100)`

**Pattern**: `\*\s*Math\.random\(\)`
- Multiplication with Math.random
- Example: `array[Math.floor(array.length * Math.random())]`

### Java
**Pattern**: `new Random\(\)`
- Non-secure Java Random (uses LCG)
- Example: `Random rand = new Random()`

**Pattern**: `Random\(\)\.next`
- Random instance method call
- Example: `new Random().nextInt()`

**Pattern**: `Random\(System\.currentTimeMillis\(\)\)`
- Time-seeded random (predictable seed)
- Example: `new Random(System.currentTimeMillis())`

**Pattern**: `\.nextInt\(`
- Random nextInt call
- Example: `rand.nextInt(100)`

**Pattern**: `\.nextLong\(`
- Random nextLong call
- Example: `rand.nextLong()`

**Pattern**: `\.nextBytes\(`
- Random nextBytes (not java.security.SecureRandom)
- Example: `rand.nextBytes(bytes)`

**Pattern**: `java\.util\.Random`
- Import of non-secure Random
- Example: `import java.util.Random`

### Go
**Pattern**: `rand\.Seed\(`
- math/rand seeding (not crypto-safe)
- Example: `rand.Seed(time.Now().UnixNano())`

**Pattern**: `rand\.Int\(`
- math/rand Int
- Example: `rand.Int()`

**Pattern**: `rand\.Intn\(`
- math/rand bounded int
- Example: `rand.Intn(100)`

**Pattern**: `rand\.Int63\(`
- math/rand 63-bit int
- Example: `rand.Int63()`

**Pattern**: `rand\.Float64\(`
- math/rand float
- Example: `rand.Float64()`

**Pattern**: `rand\.Perm\(`
- math/rand permutation
- Example: `rand.Perm(10)`

**Pattern**: `rand\.Shuffle\(`
- math/rand shuffle
- Example: `rand.Shuffle(len(slice), func(i, j int) { ... })`

**Pattern**: `rand\.Read\(`
- math/rand Read (looks like crypto/rand but isn't)
- Example: `rand.Read(buf)`

**Pattern**: `"math/rand"`
- Import of math/rand package
- Example: `import "math/rand"`

### Ruby
**Pattern**: `rand\(`
- Ruby's predictable rand (uses Mersenne Twister)
- Example: `rand(100)`

**Pattern**: `Random\.new`
- Ruby Random class (non-cryptographic)
- Example: `Random.new.rand(100)`

**Pattern**: `\.rand\(`
- Random instance method
- Example: `rng.rand(100)`

**Pattern**: `\.bytes\(`
- Random bytes (non-secure)
- Example: `Random.new.bytes(32)`

**Pattern**: `Kernel\.rand`
- Kernel rand method
- Example: `Kernel.rand(100)`

**Pattern**: `Array#sample.*random:`
- Non-secure array sampling
- Example: `array.sample(random: Random.new)`

### PHP
**Pattern**: `rand\(`
- PHP's predictable rand function (uses libc rand)
- Example: `rand(0, 100)`

**Pattern**: `mt_rand\(`
- Mersenne Twister rand (better but not crypto-safe)
- Example: `mt_rand(0, 100)`

**Pattern**: `array_rand\(`
- Non-secure array random selection
- Example: `array_rand($array)`

**Pattern**: `shuffle\(`
- Non-secure shuffle
- Example: `shuffle($array)`

**Pattern**: `str_shuffle\(`
- Non-secure string shuffle
- Example: `str_shuffle($string)`

### C/C++
**Pattern**: `\brand\(`
- C standard library rand (very weak)
- Example: `rand() % 100`

**Pattern**: `srand\(`
- Seeding weak rand
- Example: `srand(time(NULL))`

**Pattern**: `random\(`
- BSD random function
- Example: `random()`

**Pattern**: `srandom\(`
- Seeding BSD random
- Example: `srandom(time(NULL))`

**Pattern**: `drand48\(`
- 48-bit random (weak)
- Example: `drand48()`

**Pattern**: `lrand48\(`
- 48-bit random long
- Example: `lrand48()`

**Pattern**: `mrand48\(`
- 48-bit random signed
- Example: `mrand48()`

### C#
**Pattern**: `new Random\(`
- .NET Random (not cryptographically secure)
- Example: `var rand = new Random()`

**Pattern**: `Random\.Next\(`
- Random.Next method
- Example: `random.Next(100)`

**Pattern**: `Random\.NextDouble\(`
- Random.NextDouble method
- Example: `random.NextDouble()`

**Pattern**: `Random\.NextBytes\(`
- Random.NextBytes (not secure)
- Example: `random.NextBytes(buffer)`

---

## Secrets Detection

#### Hardcoded Random Seed
**Pattern**: `(?:seed|Seed|SEED)\s*[=:(]\s*(\d{1,10})\s*[);]?`
**Severity**: high
**Description**: Hardcoded seed value makes random output predictable

#### Time-Based Seed
**Pattern**: `(?:seed|Seed)\s*[=:(].*(?:time|Time|Now|now|currentTimeMillis|UnixNano)`
**Severity**: medium
**Description**: Time-based seed is predictable if attacker knows approximate time

#### Constant Seed Assignment
**Pattern**: `\bseed\s*=\s*[0-9]+\b`
**Severity**: high
**Description**: Constant seed makes all random values predictable

---

## Detection Confidence

**Import Detection**: 95%
**Usage Pattern Detection**: 90%
**Seed Pattern Detection**: 85%
