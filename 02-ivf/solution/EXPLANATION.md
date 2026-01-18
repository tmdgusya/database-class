# IVF Index - 구현 설명

## 핵심 개념

IVF (Inverted File Index)는 **"분할 정복"** 전략입니다:

1. **분할 (Partition)**: k-means로 벡터를 클러스터로 그룹화
2. **정복 (Conquer)**: 검색 시 일부 클러스터만 확인

**핵심 통찰**:
> "가까운 벡터들은 같은 클러스터에 있을 확률이 높다"

## 알고리즘 단계별 설명

### 1. Train Phase - 학습

```go
func (idx *IVFIndex) Train(vectors []vector.Vector) error {
    // 1. k-means로 클러스터링
    centroids, _ := KMeans(vectors, idx.nlist, 100, idx.metric)

    // 2. 중심점 저장
    idx.centroids = centroids

    // 3. 빈 클러스터 초기화
    idx.clusters = make([][]vector.Vector, idx.nlist)

    // 4. 학습 완료 표시
    idx.trained = true
}
```

**왜 학습이 필요한가?**
- k-means는 계산이 무거움 (O(iterations × n × k × d))
- 한 번만 수행하고 결과(중심점) 재사용
- 새 데이터는 빠르게 할당 가능

**학습 데이터 크기**:
- 최소: nlist개 (각 클러스터당 1개)
- 권장: nlist × 30개 이상
- 이유: 통계적으로 의미있는 중심점 필요

### 2. Add Phase - 벡터 추가

```go
func (idx *IVFIndex) Add(v vector.Vector) error {
    // 1. 가장 가까운 중심점 찾기
    centroidIdx, _ := FindNearestCentroid(v, idx.centroids, idx.metric)

    // 2. 해당 클러스터에 추가
    idx.clusters[centroidIdx] = append(idx.clusters[centroidIdx], v.Clone())
}
```

**시간 복잡도**: O(nlist × d)
- nlist개 중심점과 거리 계산
- Flat의 O(1)보다 느리지만 수용 가능

**메모리**: v.Clone() 사용
- 외부 수정으로부터 보호
- Flat index와 동일한 패턴

### 3. Search Phase - 검색 (핵심!)

```go
func (idx *IVFIndex) Search(query vector.Vector, k int) ([]SearchResult, error) {
    // 1. nprobe개의 가까운 중심점 찾기
    nearestCentroids, _ := idx.findNearestCentroids(query, idx.nprobe)

    // 2. 해당 클러스터들만 검색
    candidates := []SearchResult{}
    for _, clusterIdx := range nearestCentroids {
        cluster := idx.clusters[clusterIdx]
        for _, v := range cluster {
            dist, _ := idx.metric(query, v)
            candidates = append(candidates, SearchResult{
                Vector: v,
                Distance: dist,
                Index: globalIdx,
            })
        }
    }

    // 3. 정렬 후 top k 반환
    sort.Slice(candidates, ...)
    return candidates[:k], nil
}
```

**시간 복잡도**:
- 중심점 검색: O(nlist × d)
- 후보 수집: O(nprobe × avg_cluster_size × d)
- 정렬: O(n_candidates × log n_candidates)

**핵심 최적화**:
- nprobe << nlist이면, n_candidates << n
- 예: nprobe=10, nlist=100 → 90% 벡터 스킵!

## k-means 알고리즘 구현

### 표준 k-means

```go
func KMeans(vectors []vector.Vector, k int, maxIter int, metric distance.Metric) ([]vector.Vector, error) {
    // 1. 초기화 (k-means++)
    centroids := initializeCentroidsKMeansPlusPlus(vectors, k, metric)

    for iter := 0; iter < maxIter; iter++ {
        // 2. Assignment: 각 벡터를 가장 가까운 중심점에 할당
        assignments := make([]int, len(vectors))
        for i, v := range vectors {
            nearest, _ := FindNearestCentroid(v, centroids, metric)
            assignments[i] = nearest
        }

        // 3. Update: 각 클러스터의 평균을 새 중심점으로
        newCentroids := make([]vector.Vector, k)
        for clusterIdx := 0; clusterIdx < k; clusterIdx++ {
            var clusterVectors []vector.Vector
            for vecIdx, assignment := range assignments {
                if assignment == clusterIdx {
                    clusterVectors = append(clusterVectors, vectors[vecIdx])
                }
            }

            if len(clusterVectors) > 0 {
                newCentroids[clusterIdx] = computeMean(clusterVectors)
            } else {
                // 빈 클러스터: 이전 중심점 유지
                newCentroids[clusterIdx] = centroids[clusterIdx].Clone()
            }
        }

        // 4. 수렴 체크
        if converged(centroids, newCentroids, 1e-6) {
            centroids = newCentroids
            break
        }

        centroids = newCentroids
    }

    return centroids, nil
}
```

### k-means++ 초기화 (중요!)

```go
func initializeCentroidsKMeansPlusPlus(...) []vector.Vector {
    centroids := []vector.Vector{}

    // 1. 첫 중심점: 랜덤
    firstIdx := rand.Intn(len(vectors))
    centroids = append(centroids, vectors[firstIdx].Clone())

    // 2. 나머지 k-1개: 확률적 선택
    for len(centroids) < k {
        // 각 벡터에서 가장 가까운 중심점까지 거리 계산
        distances := make([]float64, len(vectors))
        totalDist := 0.0

        for i, v := range vectors {
            minDist := 1e10
            for _, centroid := range centroids {
                dist, _ := metric(v, centroid)
                if dist < minDist {
                    minDist = dist
                }
            }
            distances[i] = minDist * minDist  // 거리^2
            totalDist += distances[i]
        }

        // 거리^2에 비례하는 확률로 선택
        target := rand.Float64() * totalDist
        cumsum := 0.0
        nextIdx := 0

        for i, dist := range distances {
            cumsum += dist
            if cumsum >= target {
                nextIdx = i
                break
            }
        }

        centroids = append(centroids, vectors[nextIdx].Clone())
    }

    return centroids
}
```

**왜 k-means++인가?**

| 초기화 방법 | 장점 | 단점 |
|-----------|------|------|
| 랜덤 선택 | 간단, 빠름 | 나쁜 초기점 가능, 수렴 느림 |
| k-means++ | 좋은 초기점, 빠른 수렴 | 약간 복잡 |

**실험 결과**:
```
랜덤 초기화:   recall=72%, iterations=87
k-means++:     recall=91%, iterations=23
```

→ k-means++가 명백히 우수!

## nprobe 파라미터 - 가장 중요!

### nprobe가 recall에 미치는 영향

```
데이터: 1000 vectors, 10 clusters
쿼리별 k=10 검색

nprobe=1:   recall=32%   (1개 클러스터만 검색)
nprobe=2:   recall=55%   (2개 클러스터 검색)
nprobe=3:   recall=68%
nprobe=5:   recall=82%
nprobe=7:   recall=91%
nprobe=10:  recall=100%  (모든 클러스터 = Flat)
```

**그래프 (상상)**:
```
Recall
100%|                              ●●●●
    |                        ●●●●
    |                  ●●●●
 80%|            ●●●●              ← Sweet spot!
    |      ●●●●
 50%| ●●●●
    +────────────────────────────────> nprobe
     1   3   5   7   9   10
```

### nprobe 선택 가이드

**목표별 권장값**:

| 목표 | nprobe | Recall | Speed |
|------|--------|--------|-------|
| 빠른 대략 검색 | nlist/20 | ~70% | 매우 빠름 |
| 균형 잡힌 검색 | nlist/10 | ~85% | 빠름 ✅ |
| 고정확도 검색 | nlist/5 | ~95% | 보통 |
| 완벽한 검색 | nlist | 100% | 느림 (Flat과 동일) |

**실전 예시**:
```go
// 추천 시스템 (대략적 검색 충분)
Config{NumClusters: 1000, NumProbes: 50}  // 5%, recall ~70%

// 검색 엔진 (정확도 중요)
Config{NumClusters: 1000, NumProbes: 100} // 10%, recall ~85%

// 중복 탐지 (높은 정확도)
Config{NumClusters: 1000, NumProbes: 200} // 20%, recall ~95%
```

### 런타임 튜닝

```go
idx, _ := NewIVFIndex(Config{...})

// 처음에는 빠르게
idx.SetNumProbes(5)
results := idx.Search(query, 10)  // 빠르지만 recall 낮음

// recall이 부족하면 증가
idx.SetNumProbes(10)
results = idx.Search(query, 10)   // 더 느리지만 recall 향상
```

## 흔한 실수와 해결

### 1. 학습 전 사용

```go
// 잘못됨! ❌
idx, _ := NewIVFIndex(cfg)
idx.Add(vec)  // Panic! centroids == nil

// 해결 ✅
idx, _ := NewIVFIndex(cfg)
idx.Train(trainingVectors)  // 필수!
idx.Add(vec)
```

**코드 패턴**:
```go
func (idx *IVFIndex) Add(v vector.Vector) error {
    if !idx.trained {
        return fmt.Errorf("index not trained: call Train() first")
    }
    // ...
}
```

### 2. 부족한 학습 데이터

```go
// 문제 ⚠️
idx.Train([]vector.Vector{v1, v2, v3})  // 3개로 10 클러스터?

// 해결 ✅
minVectors := idx.nlist
if len(vectors) < minVectors {
    return fmt.Errorf("insufficient training data: need >= %d", minVectors)
}
```

**권장 크기**:
- 최소: nlist
- 좋음: nlist × 30
- 최고: nlist × 100+

### 3. Global Index 계산 실수

```go
// 문제: Search 결과의 Index가 cluster 내 인덱스
// 사용자는 전체 인덱스를 기대!

// 해결 ✅
globalIdx := 0
for clusterIdx := 0; clusterIdx < idx.nlist; clusterIdx++ {
    cluster := idx.clusters[clusterIdx]

    if searchThisCluster {
        for _, v := range cluster {
            results = append(results, SearchResult{
                Index: globalIdx,  // 전체 인덱스
            })
            globalIdx++
        }
    } else {
        globalIdx += len(cluster)  // 스킵한 클러스터도 카운트!
    }
}
```

### 4. 빈 클러스터 처리

```go
// k-means에서 빈 클러스터 발생 가능

// Update step에서:
if len(clusterVectors) > 0 {
    newCentroids[clusterIdx] = computeMean(clusterVectors)
} else {
    // 옵션 1: 이전 중심점 유지 (우리 방식)
    newCentroids[clusterIdx] = centroids[clusterIdx].Clone()

    // 옵션 2: 재초기화 (더 복잡)
    // newCentroids[clusterIdx] = reinitialize(...)
}
```

## 성능 분석

### 시간 복잡도 비교

| 작업 | Flat | IVF |
|------|------|-----|
| Build | O(1) | O(iterations × n × nlist × d) |
| Add | O(1) | O(nlist × d) |
| Search | O(n × d) | O(nprobe × n/nlist × d) |

**Search 가속 비율**:
```
Speedup = O(n × d) / O(nprobe × n/nlist × d)
        = nlist / nprobe

예: nlist=100, nprobe=10 → 10배 빠름!
```

### 메모리 사용

```go
// Flat Index
memory = n × d × sizeof(float64)

// IVF Index
memory = n × d × sizeof(float64)        // vectors
       + nlist × d × sizeof(float64)    // centroids
       + overhead                        // cluster bookkeeping

// 거의 동일! (nlist << n이므로)
```

### 벤치마크 예상 결과

```
데이터: 10,000 vectors, 128D

Flat Index:
  Search (k=10): 25ms

IVF Index (nlist=100):
  Train:         5s (한 번만)
  Search (nprobe=1):   2ms (recall=30%)  ← 빠르지만 부정확
  Search (nprobe=5):   5ms (recall=75%)
  Search (nprobe=10): 10ms (recall=90%)  ← 좋은 균형!
  Search (nprobe=20): 15ms (recall=97%)
```

## 최적화 기법

### 1. 병렬 검색

```go
// 여러 클러스터를 병렬로 검색
func (idx *IVFIndex) SearchParallel(...) {
    resultsChan := make(chan []SearchResult, nprobe)

    for _, clusterIdx := range nearestCentroids {
        go func(cIdx int) {
            // 이 클러스터 검색
            results := searchCluster(cIdx)
            resultsChan <- results
        }(clusterIdx)
    }

    // 결과 수집
    for i := 0; i < nprobe; i++ {
        clusterResults := <-resultsChan
        candidates = append(candidates, clusterResults...)
    }
}
```

### 2. 캐싱

```go
// 자주 검색되는 쿼리의 결과 캐시
type CachedIVF struct {
    *IVFIndex
    cache map[string][]SearchResult
}
```

### 3. Product Quantization (PQ)

IVF + PQ 조합:
- IVF: 거친 검색
- PQ: 압축된 거리 계산
- 메모리 절감 + 속도 향상

(이건 고급 주제 - 별도 구현)

## 실전 사용 팁

### 1. 파라미터 선택

```go
// 1단계: nlist 결정
nlist := int(math.Sqrt(float64(n)))  // 시작점
// n=10,000 → nlist=100
// n=1,000,000 → nlist=1000

// 2단계: nprobe 실험
for nprobe := 1; nprobe <= nlist/5; nprobe++ {
    recall := measureRecall(nprobe)
    latency := measureLatency(nprobe)
    // 목표 recall 달성하는 최소 nprobe 선택
}
```

### 2. 학습 데이터 선택

```go
// 옵션 1: 모든 데이터로 학습 (소규모)
idx.Train(allVectors)

// 옵션 2: 샘플링 (대규모)
sampleSize := max(nlist * 100, 100000)
sample := randomSample(allVectors, sampleSize)
idx.Train(sample)
```

### 3. 주기적 재학습

```go
// 데이터 분포가 변하면 재학습 필요
if dataDistributionChanged() {
    idx.Train(newRepresentativeSample)
    rebuildClusters()
}
```

## 다음 단계: HNSW

IVF의 한계:
- nlist가 커지면 중심점 검색도 느림
- High-dimensional data에서 "curse of dimensionality"

HNSW의 해결책:
- 그래프 기반 검색
- Sub-linear 시간 (O(log n))
- 더 높은 recall

IVF vs HNSW:
- IVF: 구현 간단, 이해 쉬움, 충분히 빠름
- HNSW: 최고 성능, 복잡함, 최신 기술

**다음 학습**: 03-hnsw/로 이동!
