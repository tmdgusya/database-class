# IVF Index - Inverted File Index

## Difficulty: 중급 (Intermediate)
**예상 학습 시간: 2-3주**

## 개요 (Overview)

IVF (Inverted File Index)는 **클러스터링**을 사용하여 검색을 가속화합니다. 핵심 아이디어는 간단합니다:

> "서로 가까운 벡터들은 같은 클러스터에 있을 것이다. 따라서 모든 벡터를 검색하지 않고, 쿼리와 가까운 클러스터만 검색하면 된다!"

**Flat Index의 문제**:
- 100만 개 벡터 → 100만 번 거리 계산 (느림!)

**IVF의 해결책**:
- 100만 개를 1000개 클러스터로 분할
- 쿼리와 가까운 10개 클러스터만 검색
- ~1000번 거리 계산 (1000배 빠름!)

**대가**: 정확도 약간 감소 (100% → 85~95%)

## 알고리즘 설명

### 1. Training Phase (학습 단계)

```
function Train(vectors):
    # k-means 알고리즘으로 클러스터링
    centroids = kmeans(vectors, num_clusters)
    store centroids
```

**한 번만 수행**: 인덱스 빌드 전에 대표 데이터로 클러스터 중심점 학습

### 2. Adding Vectors (벡터 추가)

```
function Add(vector):
    # 가장 가까운 중심점 찾기
    nearest_centroid = find_closest_centroid(vector)

    # 해당 클러스터에 할당
    assign vector to cluster[nearest_centroid]
```

**O(nlist × d)**: nlist개 중심점과 거리 계산

### 3. Searching (검색)

```
function Search(query, k):
    # nprobe개의 가까운 중심점 찾기
    nearest_centroids = find_closest_centroids(query, nprobe)

    # 해당 클러스터들만 검색
    candidates = []
    for centroid in nearest_centroids:
        candidates += vectors in cluster[centroid]

    # Top k 반환
    return k nearest from candidates
```

**핵심**: 모든 클러스터가 아닌 **nprobe개**만 검색!

## 시간 복잡도

| 작업 | Flat Index | IVF Index |
|------|-----------|-----------|
| Train | - | O(iterations × n × nlist × d) |
| Add | O(1) | O(nlist × d) |
| Search | O(n × d) | O(nprobe × avg_cluster_size × d) |

**핵심 개선**:
- nprobe << nlist이면, avg_cluster_size = n/nlist
- 검색 시간: O(nprobe × n/nlist × d) << O(n × d)
- 예: nprobe=10, nlist=100 → 10배 빠름!

## 핵심 파라미터 (매우 중요!)

### nlist (클러스터 개수)

```
nlist = 클러스터로 나눌 개수
```

**선택 기준**:
- 너무 작으면: 각 클러스터가 너무 커서 검색 느림
- 너무 크면: 학습 시간 오래 걸림, 클러스터당 벡터 너무 적음
- **권장**: `sqrt(n)` 또는 100~1000

**예시**:
- n = 10,000 → nlist = 100
- n = 1,000,000 → nlist = 1000

### nprobe (검색할 클러스터 수) ⚠️ 핵심!

```
nprobe = 검색 시 확인할 클러스터 수
```

**이것이 속도-정확도 트레이드오프의 핵심입니다!**

| nprobe | 속도 | Recall | 사용 시기 |
|--------|------|--------|----------|
| 1 | 매우 빠름 | 30-50% ⚠️ | 사용 금지 |
| 5 | 빠름 | 70-80% | 대략적 검색 |
| 10 | 보통 | 85-92% ✅ | 일반적 |
| 20 | 느림 | 92-97% | 고정확도 |
| nlist | 매우 느림 | 100% | Flat과 동일 |

**실습에서 경험할 함정**:
```bash
# nprobe=1로 테스트 실행
go test -v -run=TestIVFRecallWithSmallNprobe

# 결과: Recall = 0.32 (32%!) ❌
# 메시지: "Recall too low with nprobe=1. Try increasing nprobe!"

# nprobe=10으로 수정 후
# 결과: Recall = 0.92 (92%!) ✅
```

이것이 **가장 중요한 학습 포인트**입니다!

## k-means 알고리즘 이해

### k-means란?

**목표**: n개 벡터를 k개 그룹으로 나누기

**알고리즘**:
```
1. 초기화: k개 중심점을 랜덤하게 선택

2. 반복 (수렴할 때까지):
   a. 할당 단계 (Assignment):
      각 벡터를 가장 가까운 중심점에 할당

   b. 갱신 단계 (Update):
      각 클러스터의 중심점을 해당 클러스터 벡터들의 평균으로 갱신

3. 수렴 조건:
   - 중심점이 더 이상 변하지 않음
   - 또는 최대 반복 횟수 도달
```

### k-means++ 초기화 (더 나은 방법)

랜덤 초기화는 나쁜 클러스터를 만들 수 있습니다. k-means++는 더 좋은 초기점을 선택합니다:

```
1. 첫 중심점: 랜덤하게 선택

2. 나머지 k-1개:
   for i = 2 to k:
      # 기존 중심점에서 먼 점을 확률적으로 선택
      각 벡터 v에 대해:
          d = min_distance(v, existing_centroids)
          probability = d^2 / sum(all d^2)

      새 중심점 = sample with probability
```

**장점**: 중심점들이 서로 멀리 떨어져 더 나은 클러스터링

## 구현할 내용

### 1. kmeans.go

#### KMeans 함수
```go
func KMeans(
    vectors []vector.Vector,
    k int,
    maxIter int,
    metric distance.Metric,
) ([]vector.Vector, error)
```

**구현 단계**:
1. 초기 중심점 선택 (k-means++ 또는 랜덤)
2. 반복:
   - 각 벡터를 가장 가까운 중심점에 할당
   - 각 클러스터의 새 중심점 계산 (평균)
   - 변화 없으면 종료
3. 최종 중심점 반환

#### FindNearestCentroid 함수
```go
func FindNearestCentroid(
    v vector.Vector,
    centroids []vector.Vector,
    metric distance.Metric,
) (int, error)
```

### 2. ivf.go

#### IVFIndex 구조체
```go
type IVFIndex struct {
    centroids []vector.Vector        // 클러스터 중심점
    clusters  [][]vector.Vector       // 각 클러스터의 벡터들
    metric    distance.Metric
    nlist     int                     // 클러스터 개수
    nprobe    int                     // 검색할 클러스터 수
    trained   bool                    // 학습 여부
    dimension int
    mu        sync.RWMutex
}
```

#### 메서드들

**Train(vectors []Vector) error**
- k-means로 클러스터링
- 중심점 저장
- trained = true

**Add(v Vector) error**
- 가장 가까운 중심점 찾기
- 해당 클러스터에 추가
- **함정**: trained == false면 에러!

**Search(query Vector, k int) ([]SearchResult, error)**
- nprobe개의 가까운 중심점 찾기
- 해당 클러스터들 검색
- Top k 반환
- **함정**: nprobe가 너무 작으면 recall 저하!

**SetNumProbes(nprobe int) error**
- 런타임에 nprobe 조정
- 검증: 1 <= nprobe <= nlist

## 테스트 케이스 - 학습의 핵심!

### 기본 테스트

```go
TestIVFBasic          // 기본 동작
TestIVFNotTrained     // 학습 안 하고 사용 → 에러
TestIVFDimensionCheck // 차원 검증
```

### 🔥 함정 테스트 (가장 중요!)

#### TestIVFRecallWithSmallNprobe

```go
func TestIVFRecallWithSmallNprobe(t *testing.T) {
    // 클러스터된 데이터 생성 (1000개, 10 클러스터)
    vectors := testdata.GenerateClusteredVectors(1000, 128, 10, 42)

    // nprobe=1로 IVF 인덱스 생성 (너무 작음!)
    idx, _ := NewIVFIndex(Config{
        Metric:      distance.L2Distance,
        NumClusters: 10,
        NumProbes:   1,  // ⚠️ 함정!
    })

    idx.Train(vectors)
    for _, v := range vectors {
        idx.Add(v)
    }

    // Ground truth (Flat index)
    flatIdx := ... // 정확한 결과

    // Recall 측정
    recall := metrics.CalculateRecall(idx, flatIdx, queries, 10)

    // nprobe=1이면 recall이 매우 낮음!
    if recall < 0.6 {
        t.Errorf("❌ Recall too low with nprobe=1: %.2f%%\n"+
            "💡 Hint: Try increasing nprobe parameter!\n"+
            "   - nprobe=1: searches only 1 cluster (too few!)\n"+
            "   - nprobe=5: searches 5 clusters (better)\n"+
            "   - nprobe=10: searches all 10 clusters (perfect for this data)",
            recall*100)
    }
}
```

**학습 목표**:
- 테스트 실패 경험
- nprobe 증가로 recall 향상 확인
- 파라미터 중요성 체득

#### TestIVFRecallVsNprobe

```go
func TestIVFRecallVsNprobe(t *testing.T) {
    // nprobe를 1부터 nlist까지 변화시키며 recall 측정

    for nprobe := 1; nprobe <= 10; nprobe++ {
        idx.SetNumProbes(nprobe)
        recall := measureRecall(...)
        latency := measureLatency(...)

        fmt.Printf("nprobe=%2d: recall=%.2f%% latency=%v\n",
            nprobe, recall*100, latency)
    }

    // 예상 출력:
    // nprobe= 1: recall=32%  latency=0.5ms
    // nprobe= 2: recall=55%  latency=1.0ms
    // nprobe= 5: recall=78%  latency=2.5ms
    // nprobe=10: recall=100% latency=5.0ms
}
```

## 흔한 실수 및 함정

### 1. 학습 전 사용

```go
// 잘못됨! ❌
idx, _ := NewIVFIndex(cfg)
idx.Add(vec)  // Panic! centroids가 nil

// 올바름 ✅
idx, _ := NewIVFIndex(cfg)
idx.Train(trainingVectors)  // 먼저 학습!
idx.Add(vec)
```

### 2. 부족한 학습 데이터

```go
// 잘못됨! ❌
idx.Train([]vector.Vector{v1, v2, v3})  // 3개로 10 클러스터?

// 올바름 ✅
// 최소: nlist * 50개 권장
// 예: nlist=100 → 최소 5000개
```

### 3. nprobe = nlist (의미 없음)

```go
// 의미 없음 ⚠️
Config{
    NumClusters: 100,
    NumProbes:   100,  // 모든 클러스터 검색 = Flat index
}

// 권장 ✅
Config{
    NumClusters: 100,
    NumProbes:   10,  // 10% 검색
}
```

### 4. 너무 작은 nprobe

```go
// 테스트는 통과하지만 recall이 형편없음 ❌
Config{
    NumClusters: 100,
    NumProbes:   1,  // recall ~30%
}

// 균형잡힌 설정 ✅
Config{
    NumClusters: 100,
    NumProbes:   10,  // recall ~90%, 10배 빠름
}
```

## 구현 힌트

### k-means 구현

```go
// 간단한 k-means (k-means++ 없이)
func KMeans(vectors []vector.Vector, k int, maxIter int, metric distance.Metric) ([]vector.Vector, error) {
    // 1. 초기 중심점: 첫 k개 벡터 (또는 랜덤 샘플)
    centroids := make([]vector.Vector, k)
    for i := 0; i < k; i++ {
        centroids[i] = vectors[i].Clone()
    }

    for iter := 0; iter < maxIter; iter++ {
        // 2. 할당: 각 벡터를 가까운 중심점에 할당
        assignments := make([]int, len(vectors))
        for i, v := range vectors {
            nearest, _ := FindNearestCentroid(v, centroids, metric)
            assignments[i] = nearest
        }

        // 3. 갱신: 각 클러스터의 평균 계산
        newCentroids := computeMeans(vectors, assignments, k)

        // 4. 수렴 체크
        if converged(centroids, newCentroids) {
            break
        }

        centroids = newCentroids
    }

    return centroids, nil
}
```

### IVF Search 구현

```go
func (idx *IVFIndex) Search(query vector.Vector, k int) ([]SearchResult, error) {
    // 1. nprobe개의 가까운 중심점 찾기
    nearestCentroids := idx.findNearestCentroids(query, idx.nprobe)

    // 2. 해당 클러스터들의 모든 벡터 모으기
    candidates := []SearchResult{}
    for _, centroidIdx := range nearestCentroids {
        cluster := idx.clusters[centroidIdx]
        for vectorIdx, v := range cluster {
            dist, _ := idx.metric(query, v)
            candidates = append(candidates, SearchResult{
                Vector:   v,
                Distance: dist,
                Index:    /* global index 계산 */,
            })
        }
    }

    // 3. 정렬 후 top k 반환
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Distance < candidates[j].Distance
    })

    if k > len(candidates) {
        k = len(candidates)
    }

    return candidates[:k], nil
}
```

## 테스트 실행

```bash
cd 02-ivf/exercise

# 모든 테스트
go test -v

# 함정 테스트만
go test -v -run=TestIVFRecallWithSmallNprobe

# 벤치마크
go test -bench=. -benchmem

# nprobe별 성능 비교
go test -bench=BenchmarkIVFSearch
```

## 예상 벤치마크 결과

```
BenchmarkIVFSearch/nprobe=1-8     10000    150000 ns/op  (빠르지만 부정확)
BenchmarkIVFSearch/nprobe=5-8      5000    500000 ns/op  (균형)
BenchmarkIVFSearch/nprobe=10-8     3000   1000000 ns/op  (느리지만 정확)

vs Flat Index:
BenchmarkFlatSearch-8               500  25000000 ns/op  (25배 느림!)
```

## 학습 목표

이 실습을 완료하면:

- ✅ k-means 클러스터링 알고리즘 구현 및 이해
- ✅ 클러스터링을 통한 검색 가속화 원리
- ✅ **nprobe 파라미터의 중요성 체득** (핵심!)
- ✅ 속도-정확도 트레이드오프 경험
- ✅ Two-phase 인덱스 (Train → Use) 이해
- ✅ Flat 대비 5-10배 성능 향상 확인

## 다음 단계

IVF 완료 후:

1. ✅ Solution과 비교
2. ✅ EXPLANATION.md 읽기
3. ✅ 파라미터 튜닝 실험
4. ➡️ 03-hnsw/로 이동하여 최신 기술 학습!

## 참고 자료

- [Product Quantization for Nearest Neighbor Search](https://hal.inria.fr/inria-00514462v2/document) - IVF의 기반 논문
- [Faiss IVF 문서](https://github.com/facebookresearch/faiss/wiki/Faiss-indexes)
- k-means 애니메이션: [visualgo.net/kmeans](https://visualgo.net/en/clustering)

---

**중요**: nprobe=1로 시작해서 테스트 실패를 경험하세요. 이것이 가장 중요한 학습입니다!
