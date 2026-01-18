# Flat Index - Brute Force Search

## Difficulty: 초급 (Beginner)
**예상 학습 시간: 1주**

## 개요 (Overview)

Flat index는 가장 간단한 벡터 검색 알고리즘입니다 - 쿼리를 데이터베이스의 **모든** 벡터와 비교합니다. 느리지만 (O(n)), 다음과 같은 장점이 있습니다:

- 정확한 k-최근접 이웃을 보장 (100% recall)
- 구현과 이해가 간단함
- 다른 인덱스와 비교할 기준점

## 알고리즘 설명

```
function Search(query, k):
    distances = []
    for each vector in index:
        distance = calculate_distance(query, vector)
        distances.append((vector, distance))

    sort distances by distance
    return top k results
```

## 시간 복잡도

- **Add**: O(1) - 단순히 리스트에 추가
- **Search**: O(n × d)
  - n = 벡터 개수
  - d = 차원 수
  - 모든 벡터와 거리 계산 필요

## 공간 복잡도

- O(n × d) - 모든 벡터를 메모리에 저장

## 학습 목표

이 실습을 통해 다음을 배웁니다:

1. 벡터 간 거리 계산 방법
2. k개의 최소값 찾기 알고리즘
3. Go의 동시성 제어 (mutex)
4. 에지 케이스 처리

## 구현할 내용

`exercise/flat.go` 파일에서 다음 메서드를 구현하세요:

### 1. NewFlatIndex(cfg Config) (*FlatIndex, error)
- 인덱스 초기화
- Config 검증 (metric이 nil이 아닌지 확인)

### 2. Add(v Vector) error
- 벡터를 인덱스에 추가
- 검증 사항:
  - 벡터가 유효한가? (nil, empty, NaN, Inf 체크)
  - 차원이 일치하는가? (첫 벡터와 같은 차원)
- Thread-safety: mutex 사용

### 3. Search(query Vector, k int) ([]SearchResult, error)
- k개의 가장 가까운 벡터 찾기
- 검증 사항:
  - 쿼리 벡터가 유효한가?
  - k가 양수인가?
  - 차원이 일치하는가?
- 구현 단계:
  1. 모든 벡터와의 거리 계산
  2. k개의 최소 거리 찾기
  3. 거리순으로 정렬하여 반환
- Thread-safety: mutex 사용

### 4. Size() int
- 인덱스의 벡터 개수 반환

## 힌트

### k개 최소값 찾기

두 가지 방법이 있습니다:

**방법 1: 정렬 (간단)**
```go
// 모든 거리를 계산하고 정렬
sort.Slice(distances, func(i, j int) bool {
    return distances[i].distance < distances[j].distance
})
return distances[:k]
```

**방법 2: Heap (효율적)**
```go
// container/heap을 사용하여 k개만 유지
// 큰 데이터셋에서 더 효율적
```

Flat index의 경우 간단한 정렬로도 충분합니다. Heap은 나중에 최적화로 시도해보세요!

### Thread Safety

```go
func (idx *FlatIndex) Add(v Vector) error {
    idx.mu.Lock()         // 쓰기 잠금
    defer idx.mu.Unlock()

    // ... 구현 ...
}

func (idx *FlatIndex) Search(query Vector, k int) ([]SearchResult, error) {
    idx.mu.RLock()        // 읽기 잠금 (동시 읽기 허용)
    defer idx.mu.RUnlock()

    // ... 구현 ...
}
```

### 차원 검증

```go
// 첫 벡터일 경우 차원 저장
if idx.Size() == 0 {
    idx.dimension = v.Dimension()
}

// 차원 일치 확인
if v.Dimension() != idx.dimension {
    return fmt.Errorf("dimension mismatch: expected %d, got %d",
        idx.dimension, v.Dimension())
}
```

## 테스트 실행

```bash
cd 01-flat/exercise

# 모든 테스트 실행
go test -v

# 특정 테스트만 실행
go test -v -run=TestBasicAdd

# 벤치마크 실행
go test -bench=.

# Race condition 검사
go test -race
```

## 예상 결과

구현이 올바르면:

✅ 모든 테스트 통과
✅ Search는 정확한 k-NN 반환 (100% recall)
✅ 에지 케이스 처리 (빈 인덱스, k > size 등)
❌ 대용량 데이터에서는 느림 (괜찮습니다! 이게 baseline입니다)

## 흔한 실수

### 1. 차원 검증 누락
```go
// 잘못된 예: 차원 확인 안 함
func (idx *FlatIndex) Add(v Vector) error {
    idx.vectors = append(idx.vectors, v)
    return nil
}

// 올바른 예: 차원 확인
if idx.Size() > 0 && v.Dimension() != idx.dimension {
    return fmt.Errorf("dimension mismatch")
}
```

### 2. k > size 처리 안 함
```go
// 잘못된 예: panic 발생 가능
return results[:k]

// 올바른 예: 안전하게 처리
actualK := k
if actualK > len(results) {
    actualK = len(results)
}
return results[:actualK]
```

### 3. Mutex 잊어버리기
```go
// 잘못된 예: race condition 발생
func (idx *FlatIndex) Add(v Vector) error {
    idx.vectors = append(idx.vectors, v)  // 동시 접근 시 문제!
    return nil
}

// 올바른 예: mutex 사용
func (idx *FlatIndex) Add(v Vector) error {
    idx.mu.Lock()
    defer idx.mu.Unlock()
    idx.vectors = append(idx.vectors, v)
    return nil
}
```

### 4. 결과를 거리순으로 정렬 안 함
```go
// SearchResult는 distance 기준으로 정렬되어야 합니다!
```

## 벤치마크 이해하기

벤치마크를 실행하면 다음과 같은 결과를 볼 수 있습니다:

```
BenchmarkSearch/size=100-8      5000    250000 ns/op
BenchmarkSearch/size=1000-8      500   2500000 ns/op
BenchmarkSearch/size=10000-8      50  25000000 ns/op
```

관찰 사항:
- 벡터 개수가 10배 증가 → 시간도 약 10배 증가 (O(n) 확인)
- 이것이 **baseline**입니다
- IVF와 HNSW에서 이것보다 빨라야 합니다!

## 다음 단계

테스트가 모두 통과하면:

1. ✅ `solution/` 디렉토리의 구현과 비교
2. ✅ `EXPLANATION.md` 읽고 최적화 기법 학습
3. ✅ 벤치마크 결과 기록 (나중에 IVF/HNSW와 비교)
4. ➡️ `02-ivf/`로 이동하여 검색 가속화 학습!

## 참고 자료

- [Go sync.RWMutex 문서](https://pkg.go.dev/sync#RWMutex)
- [Go sort 패키지](https://pkg.go.dev/sort)
- [Go container/heap](https://pkg.go.dev/container/heap) (최적화용)

## 질문?

- Q: "정렬 vs Heap, 어느 것을 써야 하나요?"
  - A: 처음에는 정렬이 간단합니다. k가 n보다 훨씬 작을 때만 heap이 이점이 있습니다.

- Q: "왜 RWMutex를 쓰나요?"
  - A: Search는 읽기만 하므로 여러 goroutine이 동시에 읽을 수 있습니다. RLock을 사용하면 성능이 향상됩니다.

- Q: "실제로 이런 알고리즘을 쓰나요?"
  - A: 네! 작은 데이터셋 (<10,000 벡터)이나 100% 정확도가 필요할 때 사용합니다.
