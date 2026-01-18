# Flat Index - 구현 설명

이 문서는 Flat Index의 완전한 구현을 단계별로 설명합니다.

## 핵심 설계 결정

### 1. 데이터 구조

```go
type FlatIndex struct {
    vectors   []vector.Vector  // All stored vectors
    metric    distance.Metric  // Distance function
    dimension int              // Vector dimension (-1 = not set)
    mu        sync.RWMutex     // Thread safety
}
```

**왜 이렇게 설계했는가?**

- `vectors []vector.Vector`: 슬라이스는 동적 크기 조정이 가능하고 순차 접근에 최적
- `dimension int`: 한 번만 저장하여 모든 Add/Search에서 검증
- `sync.RWMutex`: 읽기(Search)는 동시에, 쓰기(Add)는 배타적으로

### 2. 차원 검증 전략

```go
if idx.dimension == -1 {
    // First vector - set dimension
    idx.dimension = v.Dimension()
} else {
    // Check dimension matches
    if v.Dimension() != idx.dimension {
        return fmt.Errorf("dimension mismatch: ...")
    }
}
```

**왜 -1을 사용?**
- `0`은 유효한 차원일 수 있음 (빈 벡터는 Validate에서 걸러짐)
- `-1`은 "아직 설정 안 됨"을 명확히 표현

**대안:**
```go
// 대안 1: 별도 플래그
hasDimension bool
dimension    int

// 대안 2: 첫 벡터를 확인
if len(idx.vectors) == 0 {
    // First vector...
}
```

우리 방식이 더 명확하고 간단합니다.

### 3. Thread Safety - RWMutex 사용

```go
// Add (쓰기)
func (idx *FlatIndex) Add(v vector.Vector) error {
    idx.mu.Lock()         // 배타적 잠금
    defer idx.mu.Unlock()
    // ... 수정 ...
}

// Search (읽기)
func (idx *FlatIndex) Search(...) {
    idx.mu.RLock()        // 공유 잠금
    defer idx.mu.RUnlock()
    // ... 읽기만 ...
}
```

**왜 RWMutex?**
- `Mutex`: 읽기/쓰기 모두 배타적 → Search 동시 불가
- `RWMutex`: 여러 goroutine이 동시에 Search 가능 → 성능 향상

**언제 Mutex를 쓸까?**
- 쓰기가 많고 읽기가 적을 때
- 간단함이 중요할 때

### 4. 벡터 복사 (Clone)

```go
// Store a clone to avoid external modifications
idx.vectors = append(idx.vectors, v.Clone())
```

**왜 Clone?**

잘못된 예:
```go
// 잘못됨!
idx.vectors = append(idx.vectors, v)

// 사용자 코드
vec := vector.Vector{1, 2, 3}
idx.Add(vec)
vec[0] = 999  // 인덱스 내부 벡터도 변경됨!
```

올바른 예:
```go
// 올바름
idx.vectors = append(idx.vectors, v.Clone())

vec := vector.Vector{1, 2, 3}
idx.Add(vec)
vec[0] = 999  // 인덱스는 영향 없음
```

**성능 고려:**
- Clone은 O(d) 메모리와 시간 소요
- 하지만 데이터 무결성을 위해 필수

## Search 구현 - 핵심 로직

### 1. 거리 계산

```go
results := make([]SearchResult, len(idx.vectors))
for i, v := range idx.vectors {
    dist, err := idx.metric(query, v)
    if err != nil {
        return nil, fmt.Errorf("distance calculation failed: %w", err)
    }
    results[i] = SearchResult{
        Vector:   v,
        Distance: dist,
        Index:    i,
    }
}
```

**시간 복잡도:** O(n × d)
- n번 반복 (모든 벡터)
- 각 반복에서 d번 연산 (거리 계산)

**최적화 가능?**
- Flat index는 정의상 모든 벡터를 확인해야 함
- 유일한 최적화: 병렬화 (나중에)

### 2. k-최소값 찾기

**방법 1: 전체 정렬 (우리의 구현)**

```go
sort.Slice(results, func(i, j int) bool {
    return results[i].Distance < results[j].Distance
})
return results[:k]
```

**시간 복잡도:** O(n log n)
**장점:**
- 간단하고 이해하기 쉬움
- Go 표준 라이브러리 사용
- 작은 n에서 충분히 빠름

**방법 2: Partial Sort (최적화)**

```go
// Quick select algorithm
// 시간: O(n) average, O(n^2) worst
```

**방법 3: Heap (container/heap)**

```go
import "container/heap"

// Min heap of size k
// 시간: O(n log k)
type minHeap []SearchResult

func (h minHeap) Len() int { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].Distance < h[j].Distance }
func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(SearchResult)) }
func (h *minHeap) Pop() interface{} {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[0 : n-1]
    return x
}

// Usage
h := &minHeap{}
heap.Init(h)
for _, result := range results {
    heap.Push(h, result)
    if h.Len() > k {
        heap.Pop(h)
    }
}
```

**시간 복잡도 비교:**
- 전체 정렬: O(n log n)
- Heap: O(n log k)
- Quick select: O(n) average

**언제 Heap을 쓸까?**
- k << n (예: k=10, n=1,000,000)
- 이 경우 n log k << n log n

**Flat index에서는?**
- 대부분 k와 n이 작음 (< 10,000)
- 정렬이 충분히 빠르고 간단함
- 큰 데이터셋은 IVF/HNSW 사용

### 3. k > size 처리

```go
// Return top k (or all if k > size)
if k > len(results) {
    k = len(results)
}
return results[:k]
```

**왜 에러가 아니라 조정?**

사용자 관점에서:
```go
// "최대 10개까지 주세요"
results, _ := idx.Search(query, 10)

// 인덱스에 5개만 있으면?
// 옵션 1: 에러 (사용자가 처리 필요)
// 옵션 2: 5개 반환 (우리 방식, 편리함)
```

다른 라이브러리:
- Faiss: k 조정
- Annoy: k 조정
- HNSW: k 조정

**일관성**: 대부분의 라이브러리와 동일한 동작

## 에러 처리 철학

### 입력 검증

```go
// Validate query
if err := query.Validate(); err != nil {
    return nil, fmt.Errorf("invalid query: %w", err)
}

if k <= 0 {
    return nil, fmt.Errorf("k must be positive, got %d", k)
}
```

**원칙:**
1. **조기 검증**: 작업 전에 모든 입력 확인
2. **명확한 메시지**: 무엇이 잘못되었는지 설명
3. **Error wrapping**: `%w`로 원본 에러 유지

### Panic vs Error

**Panic을 쓰면 안 되는 경우:**
```go
// 잘못됨!
if v.Dimension() != idx.dimension {
    panic("dimension mismatch")
}
```

**Error를 반환해야 하는 이유:**
- 사용자 코드에서 복구 가능
- 테스트에서 확인 가능
- 서버가 죽지 않음

**Panic을 써도 되는 경우:**
- 프로그래머 실수 (예: "not implemented")
- 복구 불가능한 상황 (예: 메모리 부족)

## 성능 고려사항

### 메모리 할당

```go
// 미리 크기 할당
results := make([]SearchResult, len(idx.vectors))

// vs 동적 추가
results := []SearchResult{}
results = append(results, ...)  // 여러 번 재할당
```

**우리 방식이 빠른 이유:**
- 한 번의 할당
- 재할당 없음
- 캐시 지역성 향상

### 벤치마크 결과 예상

```
BenchmarkSearch/size=100-8      5000    250000 ns/op    10000 B/op    10 allocs/op
BenchmarkSearch/size=1000-8      500   2500000 ns/op   100000 B/op   100 allocs/op
BenchmarkSearch/size=10000-8      50  25000000 ns/op  1000000 B/op  1000 allocs/op
```

**관찰:**
- 시간: O(n) 스케일링
- 메모리: O(n) 증가
- Allocs: 각 SearchResult 할당

## 테스트 커버리지

### 중요 테스트 케이스

1. **기본 기능**
   - Add, Search, Size
   - 정확한 거리 계산
   - 정렬 순서

2. **에지 케이스**
   - 빈 인덱스
   - k > size
   - 중복 벡터
   - 차원 불일치

3. **에러 케이스**
   - nil, empty, NaN, Inf
   - nil metric

4. **동시성**
   - 동시 Add
   - 동시 Search
   - 혼합

### Race Detector

```bash
go test -race
```

**발견할 수 있는 문제:**
- Mutex 누락
- Shared memory 접근
- Data race

## 다음 단계

### IVF Index (02-ivf/)

Flat index의 문제:
- O(n) 시간 → 느림
- 모든 벡터 확인 → 비효율

IVF의 해결책:
- 클러스터링 → O(nprobe × cluster_size)
- 일부 클러스터만 검색
- Trade-off: 속도 vs 정확도

### 학습한 패턴

1. Thread-safe 설계 (RWMutex)
2. 입력 검증 (Validate)
3. 에러 처리 (wrapping)
4. 메모리 최적화 (미리 할당)

이 패턴들은 IVF와 HNSW에서도 사용됩니다!
