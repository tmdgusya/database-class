# 벡터 검색 이론 (Vector Search Theory)

## ANN (Approximate Nearest Neighbors)이란?

### 문제 정의

**k-Nearest Neighbors (k-NN)**:
> 주어진 쿼리 벡터 q와 데이터베이스 D에서 가장 가까운 k개 벡터 찾기

**정확한 k-NN** (Exact):
```
정의: dist(q, v1) <= dist(q, v2) <= ... <= dist(q, vk)
시간: O(n × d)  // 모든 벡터 확인 필요
```

**근사 k-NN** (Approximate):
```
정의: "거의" 가장 가까운 k개 반환
시간: O(log n) 또는 O(sqrt(n))  // 훨씬 빠름!
대가: Recall < 100%
```

### Recall 정의

```
Recall@k = (정확한 k-NN과의 교집합 크기) / k

예:
정확한 k-NN: [v1, v2, v3, v4, v5]
ANN 결과:    [v1, v2, v6, v7, v8]
교집합:      [v1, v2]
Recall@5:    2/5 = 40%
```

**목표**: Recall 90-95% + 빠른 속도

## Curse of Dimensionality (차원의 저주)

### 문제

고차원 공간에서:
1. **모든 점이 멀어짐**
   ```
   2D:   최근접과 최원접의 거리 비율 ~2:1
   10D:  거리 비율 ~1.5:1
   100D: 거리 비율 ~1.1:1  ← 거의 차이 없음!
   ```

2. **공간이 희박해짐 (Sparse)**
   ```
   단위 구의 부피:
   10D:  0.0025  (구석에 대부분 밀집)
   100D: 10^-40  (거의 비어있음)
   ```

3. **모든 벡터가 "비슷하게" 보임**
   - 거리 차이가 미미
   - k-NN이 의미 없어짐

### 해결책

1. **차원 축소**
   - PCA, t-SNE, UMAP
   - 100D → 10D

2. **좋은 인덱스 구조**
   - HNSW, IVF
   - 고차원에 강건한 알고리즘

3. **학습된 거리**
   - Metric learning
   - 의미있는 거리 함수 학습

## 주요 알고리즘 비교

### 트리 기반 (Tree-based)

#### KD-Tree
```
구조: Binary space partitioning
시간: O(log n) in low-dim, O(n) in high-dim
문제: 고차원에서 퇴화
사용: < 20D에서만 효과적
```

#### Ball Tree
```
구조: Metric tree
시간: O(log n) in low-dim
장점: KD-tree보다 고차원에 강함
문제: 여전히 > 50D에서 느림
```

**결론**: 트리는 고차원에서 실패

### 해시 기반 (Hash-based)

#### LSH (Locality Sensitive Hashing)
```
아이디어: 가까운 벡터를 같은 버킷에 해시
시간: O(1) per query
장점: 매우 빠름
단점: Recall이 불안정 (60-80%)
사용: 초고속 대략 검색
```

**구현 원리**:
```python
# Random projection
for i in range(L):  # L hash tables
    random_vectors = generate_random(d, k)
    hash_values[i] = dot(query, random_vectors) > 0
```

### 클러스터 기반 (Clustering-based)

#### IVF (우리가 구현한 것!)
```
아이디어: k-means로 공간 분할, 일부만 검색
시간: O(nprobe × cluster_size)
장점: 구현 간단, 안정적 recall
파라미터: nlist, nprobe
사용: 중소규모 시스템
```

#### IVF-PQ
```
개선: Product Quantization 추가
메모리: 10-100배 절감
사용: Faiss의 기본 인덱스
```

### 그래프 기반 (Graph-based)

#### NSW (Navigable Small World)
```
아이디어: Small world graph 탐색
시간: O(log n)
문제: 빌드가 느림 (O(n^2))
```

#### HNSW (우리가 구현한 것!)
```
개선: 계층 구조 추가 → 빌드 O(n log n)
시간: O(log n) search
장점: 최고 성능, 안정적
단점: 메모리 많이 씀
사용: 현재 SOTA (state-of-the-art)
```

## 거리 메트릭 (Distance Metrics)

### L2 (Euclidean) Distance

```
d(a, b) = sqrt(sum((a[i] - b[i])^2))
```

**특성**:
- 절대적 거리
- Triangle inequality 만족
- 이미지, 오디오에 적합

**문제**:
- 고차원에서 모든 거리가 비슷
- 크기에 민감 (정규화 필요)

### Cosine Distance

```
cos_sim(a, b) = dot(a, b) / (||a|| × ||b||)
cos_dist(a, b) = 1 - cos_sim(a, b)
```

**특성**:
- 방향만 고려, 크기 무시
- [-1, 1] → [0, 2] 범위
- 텍스트 임베딩에 최적

**장점**:
- 정규화 자동
- 고차원에서 강건

### Inner Product (Dot Product)

```
ip(a, b) = sum(a[i] × b[i])
```

**특성**:
- 크기와 방향 모두 고려
- Metric이 아님 (triangle inequality 불만족)
- ML 모델 출력에 적합

**사용**:
- 정규화된 벡터: Cosine과 동일
- Matrix factorization

## Recall-Speed-Memory Trade-off

### 3차원 트레이드오프

```
        High Recall
            △
           /|\
          / | \
         /  |  \
        /   |   \
       /  HNSW  \
      /     |     \
     /    IVF      \
    /       |       \
   /      Flat       \
  /         |         \
 /__________o__________\
Fast                   Low
Speed                  Memory
```

### 알고리즘 포지셔닝

| 알고리즘 | Recall | Speed | Memory | 복잡도 |
|---------|--------|-------|--------|--------|
| Flat | 100% | 느림 | 낮음 | 쉬움 |
| LSH | 60-80% | 매우빠름 | 중간 | 보통 |
| IVF | 85-95% | 빠름 | 낮음 | 보통 |
| HNSW | 95-99% | 매우빠름 | 높음 | 어려움 |

### 선택 가이드

**Flat**: n < 10K, 100% recall 필요
**IVF**: 10K < n < 1M, 균형
**HNSW**: n > 100K, 최고 성능
**LSH**: 초고속, recall 타협 가능

## Quantization (양자화)

### Scalar Quantization

```python
# float32 → uint8
min_val, max_val = compute_range(vectors)
quantized = ((vectors - min_val) / (max_val - min_val) * 255).astype(uint8)

# 메모리: 4배 절감
# 정확도: 약간 손실 (<1% recall 감소)
```

### Product Quantization (PQ)

```python
# 128D → 8 subvectors of 16D
# 각 subvector를 256개 centroids로 양자화
# 128 × 4bytes → 8 × 1byte = 32배 절감!

# IVF + PQ = Faiss의 기본
```

**트레이드오프**:
- 메모리 ↓↓
- 속도 ↑ (캐시 효율)
- Recall ↓ (약간)

## 실전 시스템 설계

### 스케일별 권장 구조

**소규모 (< 100K)**:
```
Flat 또는 IVF
- 간단, 충분히 빠름
- 메모리 효율
```

**중규모 (100K - 10M)**:
```
IVF + PQ 또는 HNSW
- 안정적 성능
- 파라미터 튜닝 필요
```

**대규모 (> 10M)**:
```
HNSW + Quantization
+ Distributed index
+ GPU acceleration
```

### 최적화 기법

1. **Batch Processing**
   ```go
   // 개별 쿼리보다 배치가 빠름
   results := idx.BatchSearch(queries, k)
   ```

2. **Pre-filtering**
   ```go
   // 메타데이터로 먼저 필터링
   candidates := filter_by_category(category)
   results := idx.Search(query, k, candidates)
   ```

3. **Caching**
   ```go
   // 빈번한 쿼리 캐싱
   if cached := cache.Get(query); cached != nil {
       return cached
   }
   ```

4. **GPU Acceleration**
   - 거리 계산 병렬화
   - 100배+ 속도 향상

## Further Reading

### 논문
- HNSW: [Efficient and robust approximate nearest neighbor search](https://arxiv.org/abs/1603.09320)
- IVF: [Product Quantization for Nearest Neighbor Search](https://hal.inria.fr/inria-00514462)
- LSH: [Locality-Sensitive Hashing Scheme Based on p-Stable Distributions](https://www.cs.princeton.edu/courses/archive/spr05/cos598E/bib/LSH.pdf)

### 라이브러리
- Faiss (Meta): https://github.com/facebookresearch/faiss
- Annoy (Spotify): https://github.com/spotify/annoy
- hnswlib: https://github.com/nmslib/hnswlib

### 벤치마크
- ann-benchmarks.com: 알고리즘 성능 비교

---

이론을 실전에 적용하려면 직접 구현해보는 것이 최고입니다. 이 코스를 완료하면 모든 핵심 개념을 체득하게 됩니다!
