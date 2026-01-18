# HNSW - Hierarchical Navigable Small World

## Difficulty: 고급 (Advanced)
**예상 학습 시간: 4-6주**

## 개요 (Overview)

HNSW (Hierarchical Navigable Small World)는 현재 **가장 인기있는** ANN (Approximate Nearest Neighbor) 알고리즘입니다.

**핵심 아이디어**: "고속도로 시스템"
- **상위 레이어**: 고속도로 - 적은 연결, 큰 점프
- **하위 레이어**: 일반 도로 - 많은 연결, 작은 점프
- **검색**: 고속도로로 빠르게 접근 → 일반 도로로 정밀 탐색

### 왜 HNSW인가?

| 알고리즘 | 검색 시간 | Recall | 메모리 | 복잡도 |
|---------|----------|--------|--------|--------|
| Flat | O(n) | 100% | 낮음 | ⭐ 쉬움 |
| IVF | O(n/nlist) | 85-95% | 낮음 | ⭐⭐ 보통 |
| **HNSW** | **O(log n)** | **95-99%** | 높음 | ⭐⭐⭐ 어려움 |

**실전 성능**:
- Faiss, Milvus, Weaviate 등 주요 벡터 DB의 기본 인덱스
- 수백만~수십억 벡터에서 1ms 이내 검색
- Recall 95%+ 달성

## 알고리즘 직관

### Skip List + Proximity Graph

HNSW = **Probabilistic Skip List** + **Navigable Small World Graph**

#### Skip List (건물 비유)

```
Level 3 (꼭대기): 1 -----> 10 -----> 100      (소수만 연결)
Level 2:         1 ---> 5 ---> 10 ---> 50 ---> 100
Level 1:         1 -> 3 -> 5 -> 7 -> 10 -> ... -> 100
Level 0 (바닥):  1-2-3-4-5-6-7-8-9-10-...-100   (모두 연결)
```

**검색**:
1. 꼭대기에서 시작
2. 현재 레벨에서 최대한 가깝게 이동
3. 더 이상 못 가면 한 층 내려감
4. 반복

**시간**: O(log n) - 각 층마다 절반씩 좁혀짐

#### Proximity Graph (친구 네트워크)

벡터 공간에서:
- 각 노드가 가까운 이웃들과 연결
- 그래프 탐색으로 검색
- "친구의 친구"를 따라가면 목표 도달

**Small World 속성**:
- 평균 6 hop으로 모든 노드 도달 가능 (6 degrees of separation)
- 클러스터링 + Long-range 연결

### HNSW = 두 개념의 결합

```
Layer 3:  A =================> Z          (Long-range jumps)
          ↓                    ↓
Layer 2:  A ======> M ======> Z           (Medium jumps)
          ↓         ↓         ↓
Layer 1:  A -> B -> M -> ... -> Z        (Short jumps)
          ↓    ↓    ↓         ↓
Layer 0:  A-B-C-...-M-...-Y-Z             (All vectors)

각 노드는 가까운 이웃들과 연결 (M개)
```

## 알고리즘 상세

### 1. 데이터 구조

```go
type HNSWIndex struct {
    nodes      []Node           // 모든 노드
    entryPoint int              // 최상위 레이어의 진입점
    maxLayer   int              // 현재 최대 레이어
    M          int              // 레이어당 최대 연결 수
    efConstruction int          // 빌드 시 탐색 크기
    efSearch   int              // 검색 시 탐색 크기
    metric     distance.Metric
}

type Node struct {
    ID          int
    Vector      vector.Vector
    Connections [][]int  // connections[layer] = neighbor IDs
    Level       int      // 이 노드의 최대 레이어
}
```

### 2. Random Level Generation

각 노드가 어느 레이어까지 올라갈지 확률적 결정:

```go
func RandomLevel(ml float64) int {
    level := 0
    for rand.Float64() < ml && level < maxLevel {
        level++
    }
    return level
}

// ml = 1/ln(2) ≈ 0.69
// P(level=0) = 50%
// P(level=1) = 25%
// P(level=2) = 12.5%
// ...
```

**결과**: 피라미드 구조
```
Layer 4: 1 node
Layer 3: ~2 nodes
Layer 2: ~4 nodes
Layer 1: ~8 nodes
Layer 0: 100 nodes
```

### 3. Insertion Algorithm

```python
function Add(newVector):
    # 1. 레벨 생성
    level = RandomLevel()

    # 2. Entry point에서 시작하여 layer-by-layer 검색
    currNearest = [entryPoint]

    # 3. 상위 레이어들 (greedy search만)
    for layer from maxLayer down to level+1:
        currNearest = searchLayer(newVector, currNearest, ef=1, layer)

    # 4. level부터 0까지 (삽입 + 연결)
    for layer from level down to 0:
        # 후보 찾기
        candidates = searchLayer(newVector, currNearest, efConstruction, layer)

        # M개의 이웃 선택
        neighbors = selectNeighbors(candidates, M, layer)

        # 양방향 연결
        connect(newVector, neighbors, layer)

        # 이웃들의 연결 개수가 M 초과하면 pruning
        for neighbor in neighbors:
            if len(neighbor.connections[layer]) > M:
                prune(neighbor, layer, M)

        currNearest = neighbors

    # 5. Entry point 업데이트 (필요시)
    if level > maxLayer:
        entryPoint = newVector
        maxLayer = level
```

### 4. Search Algorithm

```python
function Search(query, k):
    # 1. Entry point에서 시작
    currNearest = [entryPoint]

    # 2. 상위 레이어들 (빠른 탐색)
    for layer from maxLayer down to 1:
        currNearest = searchLayer(query, currNearest, ef=1, layer)

    # 3. Layer 0 (정밀 탐색)
    candidates = searchLayer(query, currNearest, efSearch, layer=0)

    # 4. Top k 반환
    return top k from candidates
```

### 5. searchLayer (핵심!)

```python
function searchLayer(query, entryPoints, ef, layer):
    # ef: 탐색할 후보 수 (작으면 빠르지만 부정확, 크면 느리지만 정확)

    visited = set()
    candidates = min-heap(entryPoints)  # 거리 기준 min heap
    best = max-heap(entryPoints)        # 거리 기준 max heap (상위 ef개 유지)

    while candidates not empty:
        current = pop closest from candidates

        if current.distance > best.worst_distance:
            break  # 더 이상 개선 불가

        # 현재 노드의 이웃들 확인
        for neighbor in current.connections[layer]:
            if neighbor in visited:
                continue

            visited.add(neighbor)
            dist = distance(query, neighbor.vector)

            if dist < best.worst_distance or len(best) < ef:
                push neighbor to candidates
                push neighbor to best

                if len(best) > ef:
                    pop worst from best

    return best (ef개 노드)
```

## 핵심 파라미터 (매우 중요!)

### M - Maximum connections per layer

```
M = 각 노드가 가질 수 있는 최대 이웃 수
```

**영향**:
- **M ↑**:
  - 그래프 연결성 ↑ → Recall ↑
  - 메모리 ↑ (O(n × M))
  - 빌드 시간 ↑
  - 검색 시간 약간 ↑ (더 많은 이웃 확인)

- **M ↓**:
  - 메모리 ↓
  - 그래프 단절 가능 → Recall ↓↓↓

**권장값**:
- Low-dim (< 100D): M = 12-16
- High-dim (> 100D): M = 32-48

**함정**:
```go
Config{M: 4}  // ❌ 너무 작음! Recall < 70%
Config{M: 16} // ✅ 균형 잡힘
Config{M: 64} // ⚠️  메모리 많이 쓰지만 recall 최고
```

### efConstruction - Construction-time candidate list size

```
efConstruction = 빌드 시 searchLayer의 ef 값
```

**영향**:
- **efConstruction ↑**:
  - 그래프 품질 ↑ → Recall ↑
  - 빌드 시간 ↑↑ (큰 영향!)

- **efConstruction ↓**:
  - 빌드 빠름
  - 그래프 품질 ↓ → Recall ↓

**권장값**:
- Quick build: efConstruction = 100
- Balanced: efConstruction = 200
- High quality: efConstruction = 400

**빌드 시간 예상**:
```
efConstruction=100:  10s, recall=85%
efConstruction=200:  30s, recall=93% ✅
efConstruction=400: 120s, recall=96%
```

### efSearch - Search-time candidate list size

```
efSearch = 검색 시 searchLayer의 ef 값
```

**영향**:
- **efSearch ↑**:
  - Recall ↑
  - 검색 시간 ↑

- **efSearch ↓**:
  - 검색 빠름
  - Recall ↓

**중요 제약**:
```
efSearch >= k  (필수!)
```

**권장값**:
- Fast search: efSearch = k
- Balanced: efSearch = 2k - 4k
- High recall: efSearch = 10k

**런타임 조정 가능**:
```go
idx.SetEfSearch(50)  // 빠른 검색
results := idx.Search(query, 10)

idx.SetEfSearch(200) // 정확한 검색
results = idx.Search(query, 10)
```

## 구현할 내용

### 1. node.go

```go
type Node struct {
    ID          int
    Vector      vector.Vector
    Connections [][]int  // connections[layer] = neighbor IDs
    Level       int      // Max level of this node
}

func NewNode(id int, v vector.Vector, level int) *Node
```

### 2. layer.go

```go
// RandomLevel generates random level with exponential decay
func RandomLevel(ml float64, maxLevel int) int
```

### 3. hnsw.go

```go
type HNSWIndex struct {
    // TODO: Add fields
    // - nodes []Node
    // - entryPoint int
    // - maxLayer int
    // - M, Mmax, efConstruction, efSearch int
    // - metric
    // - mu
}

func NewHNSWIndex(cfg Config) (*HNSWIndex, error)

func (idx *HNSWIndex) Add(v vector.Vector) error {
    // 복잡한 삽입 알고리즘!
}

func (idx *HNSWIndex) Search(query vector.Vector, k int) ([]SearchResult, error) {
    // Layer-by-layer search
}

func (idx *HNSWIndex) searchLayer(...) []nodeWithDistance {
    // Greedy search with visited set
}

func (idx *HNSWIndex) selectNeighbors(...) []int {
    // Select M best neighbors
    // Simple: closest M
    // Advanced: heuristic (diversity)
}
```

## 테스트 케이스 - 학습의 핵심!

### 기본 테스트

```go
TestHNSWBasic          // 기본 add/search
TestHNSWLevels         // Random level 분포
TestHNSWConnectivity   // 그래프 연결성
```

### 🔥 함정 테스트

#### TestHNSWPoorConnectivity

```go
// M=2 (너무 작음!)
idx, _ := NewHNSWIndex(Config{
    M:              2,  // ❌ 함정!
    efConstruction: 100,
    efSearch:       50,
})

// Recall이 매우 낮음 (<70%)
// 이유: 그래프가 단절됨
```

#### TestHNSWEfSearchTooSmall

```go
// efSearch < k
idx, _ := NewHNSWIndex(Config{
    M:              16,
    efConstruction: 200,
    efSearch:       5,   // ❌ 함정!
})

results, _ := idx.Search(query, 10)  // k=10인데 efSearch=5?
// Error 또는 5개만 반환
```

#### TestHNSWParameterSweep

```go
// M, efConstruction, efSearch 조합 실험
configs := []Config{
    {M: 4,  efC: 40,  efS: 10},  // 나쁨
    {M: 16, efC: 200, efS: 50},  // 좋음 ✅
    {M: 32, efC: 400, efS: 100}, // 최고 (느림)
}

for _, cfg := range configs {
    // recall, build time, search time 측정
}
```

## 흔한 실수 및 함정

### 1. efSearch < k

```go
// 잘못됨! ❌
idx.SetEfSearch(5)
results, _ := idx.Search(query, 10)  // k=10 요청

// 문제: efSearch는 탐색할 후보 수
//      후보가 5개인데 10개를 어떻게 반환?

// 해결 ✅
if efSearch < k {
    return error("efSearch must be >= k")
}
```

### 2. M 너무 작음

```go
// ❌ M=2는 재앙
Config{M: 2}

// 결과:
// - 그래프가 여러 조각으로 단절
// - Entry point에서 일부 노드 도달 불가
// - Recall < 60%

// ✅ 최소 M=8, 권장 M=16
```

### 3. Entry Point 업데이트 누락

```go
// Add에서:
if newNode.Level > idx.maxLayer {
    idx.entryPoint = newNode.ID  // 필수!
    idx.maxLayer = newNode.Level
}

// 누락하면:
// - 높은 레이어에 노드가 있는데 entry point가 낮음
// - 일부 노드 도달 불가
```

### 4. Visited Set 없이 searchLayer

```go
// ❌ 무한 루프!
func searchLayer(...) {
    for !candidates.Empty() {
        curr := candidates.Pop()
        for _, neighbor := range curr.Connections {
            candidates.Push(neighbor)  // 방문 체크 안 함!
        }
    }
}

// ✅ Visited set 사용
visited := make(map[int]bool)
```

### 5. 양방향 연결 누락

```go
// ❌ 단방향만
newNode.Connections[layer] = append(..., neighborID)

// ✅ 양방향
newNode.Connections[layer] = append(..., neighborID)
neighbor.Connections[layer] = append(..., newNode.ID)
```

## 구현 힌트

### searchLayer 구현

```go
type nodeWithDistance struct {
    nodeID   int
    distance float64
}

func (idx *HNSWIndex) searchLayer(
    query vector.Vector,
    entryPoints []int,
    ef int,
    layer int,
) []nodeWithDistance {
    // Min heap for candidates (탐색할 노드들)
    candidates := &minHeap{}

    // Max heap for best results (상위 ef개 유지)
    best := &maxHeap{}

    visited := make(map[int]bool)

    // Initialize
    for _, ep := range entryPoints {
        dist, _ := idx.metric(query, idx.nodes[ep].Vector)
        heap.Push(candidates, nodeWithDistance{ep, dist})
        heap.Push(best, nodeWithDistance{ep, dist})
        visited[ep] = true
    }

    for candidates.Len() > 0 {
        curr := heap.Pop(candidates).(nodeWithDistance)

        // Early termination
        if curr.distance > best.Peek().distance {
            break
        }

        // Explore neighbors
        for _, neighborID := range idx.nodes[curr.nodeID].Connections[layer] {
            if visited[neighborID] {
                continue
            }
            visited[neighborID] = true

            dist, _ := idx.metric(query, idx.nodes[neighborID].Vector)

            if dist < best.Peek().distance || best.Len() < ef {
                heap.Push(candidates, nodeWithDistance{neighborID, dist})
                heap.Push(best, nodeWithDistance{neighborID, dist})

                if best.Len() > ef {
                    heap.Pop(best)
                }
            }
        }
    }

    // Return ef best nodes
    results := make([]nodeWithDistance, best.Len())
    for i := best.Len() - 1; i >= 0; i-- {
        results[i] = heap.Pop(best).(nodeWithDistance)
    }

    return results
}
```

### selectNeighbors 구현

**Simple 버전** (처음 구현):
```go
func (idx *HNSWIndex) selectNeighbors(
    candidates []nodeWithDistance,
    M int,
) []int {
    // 단순히 가장 가까운 M개 선택
    if len(candidates) <= M {
        result := make([]int, len(candidates))
        for i, c := range candidates {
            result[i] = c.nodeID
        }
        return result
    }

    // Sort by distance
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].distance < candidates[j].distance
    })

    result := make([]int, M)
    for i := 0; i < M; i++ {
        result[i] = candidates[i].nodeID
    }
    return result
}
```

**Advanced 버전** (최적화):
```go
// Heuristic: 가까움 + 다양성
// 너무 밀집된 이웃들 대신 공간적으로 분산된 이웃 선호
```

## 성능 목표

### 예상 결과 (10,000 vectors, 128D)

| 작업 | Flat | IVF | HNSW |
|------|------|-----|------|
| Build | 즉시 | 5s | 30s |
| Search (k=10) | 25ms | 2.5ms | 0.3ms |
| Recall | 100% | 90% | 97% |
| Memory | 10MB | 10MB | 15MB |

### 벤치마크 출력 예시

```
HNSW (M=16, efC=200, efS=50):
  Build:  30.2s
  Search: 0.31ms/query
  Recall: 96.8%
  QPS:    3200 queries/sec
```

## 학습 목표

이 실습을 완료하면:

- ✅ 그래프 기반 검색 알고리즘 이해
- ✅ 계층적 구조의 장점 체득
- ✅ 복잡한 파라미터 상호작용 경험
- ✅ Sub-linear 검색 시간 달성
- ✅ Heap, Priority Queue 활용
- ✅ 현재 최고 수준의 ANN 기술 습득

## 다음 단계

HNSW 완료 후:

1. ✅ Solution 비교
2. ✅ EXPLANATION.md 읽기
3. ✅ 파라미터 실험 (M, efC, efS)
4. ✅ 세 인덱스 종합 비교
5. ➡️ examples/로 이동하여 실전 활용!

## 참고 자료

- [Original HNSW Paper](https://arxiv.org/abs/1603.09320) - Malkov & Yashunin, 2016
- [hnswlib](https://github.com/nmslib/hnswlib) - 공식 C++ 구현
- [HNSW Explained](https://www.pinecone.io/learn/hnsw/) - Pinecone 블로그

---

**중요**: 파라미터 실험을 많이 해보세요! HNSW의 진가는 올바른 파라미터에서 나옵니다.
