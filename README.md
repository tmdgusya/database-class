# Vector Database Indexing Course in Go

**벡터 데이터베이스의 핵심 인덱싱 기법을 from scratch로 구현하며 배우는 실습 과정**

## 🎯 무엇을 배우나요?

이 코스에서는 세 가지 핵심 벡터 인덱스를 직접 구현하며 학습합니다:

1. **Flat Index** (1주차) - 브루트포스 기준점
   - O(n) 선형 검색
   - 100% 정확도 보장
   - 다른 인덱스의 baseline

2. **IVF Index** (2-3주차) - 클러스터 기반 가속
   - k-means 클러스터링 구현
   - nprobe 파라미터를 통한 속도/정확도 트레이드오프
   - 실전에서 가장 많이 사용되는 기법 중 하나

3. **HNSW Index** (4-6주차) - 최신 그래프 기반 인덱스
   - 계층적 skip list 구조
   - Sub-linear 검색 시간
   - 현재 가장 인기있는 ANN 알고리즘

## 🏗️ 학습 철학

### 제공되는 것
- ✅ 완전한 유틸리티 라이브러리 (pkg/)
- ✅ 구조체 정의 및 메서드 시그니처 (skeleton)
- ✅ 포괄적인 테스트 케이스 (함정 포함!)
- ✅ 상세한 이론 설명 (README)
- ✅ 완전한 답안 코드 (solution/)

### 직접 구현할 것
- 🔨 핵심 알고리즘 (Add, Search, 클러스터링, 그래프 탐색)
- 🔨 데이터 구조 설계
- 🔨 파라미터 튜닝 및 최적화

### 특별한 학습 요소: **의도적 함정 (Traps)**

테스트 케이스에 의도적인 함정이 숨어있습니다:

- ❌ **IVF with nprobe=1**: Recall이 30%로 떨어짐 → nprobe 증가 필요성 체감
- ❌ **HNSW with efSearch<k**: 런타임 에러 → 파라미터 제약조건 학습
- ❌ **차원 불일치**: Add/Search 실패 → 입력 검증의 중요성

**목표**: 실패를 통해 배우고, 파라미터 튜닝의 중요성을 몸소 체험!

## 📁 프로젝트 구조

```
vector-db-class/
├── README.md                      # 이 파일
├── LEARNING_PATH.md               # 상세 학습 로드맵
│
├── pkg/                           # 공유 유틸리티 (완전 구현됨)
│   ├── vector/                   # Vector 타입 및 연산
│   ├── distance/                 # L2, Cosine, Dot 거리 메트릭
│   ├── testdata/                 # 테스트 데이터 생성기
│   └── metrics/                  # Recall 및 성능 측정
│
├── 01-flat/                       # Week 1: Brute Force
│   ├── README.md                 # 이론 및 구현 가이드
│   ├── exercise/                 # 여기서 구현!
│   │   ├── flat.go               # TODO: 구현 필요
│   │   └── flat_test.go          # 테스트 케이스
│   └── solution/                 # 참고용 완전 구현
│       ├── flat.go
│       └── EXPLANATION.md        # 상세 설명
│
├── 02-ivf/                        # Week 2-3: IVF (계획 참고)
│   ├── README.md
│   ├── exercise/
│   └── solution/
│
├── 03-hnsw/                       # Week 4-6: HNSW (계획 참고)
│   ├── README.md
│   ├── exercise/
│   └── solution/
│
├── examples/                      # 실전 예제 (계획 참고)
│   ├── 01-basic-search/
│   ├── 02-parameter-tuning/
│   └── 03-recall-vs-speed/
│
├── docs/                          # 학습 자료 (계획 참고)
│   ├── THEORY.md
│   ├── BENCHMARKING.md
│   ├── COMMON_PITFALLS.md
│   └── FURTHER_READING.md
│
└── scripts/                       # 편의 스크립트 (계획 참고)
    ├── run_all_tests.sh
    └── benchmark_all.sh
```

## 🚀 시작하기

### 필수 요구사항

- **Go 1.19+** 설치
- Git
- 기본적인 Go 문법 지식
- 선형대수 기초 (벡터, 거리 개념)

### 설치

```bash
# 저장소 클론 (또는 현재 디렉토리 사용)
cd vector-db-class

# 의존성 다운로드
go mod download

# pkg 패키지 테스트 (모두 통과해야 함)
cd pkg/vector && go test -v
cd ../distance && go test -v
cd ../..
```

### Quick Start (5분)

```bash
# 1. Flat index README 읽기
cat 01-flat/README.md

# 2. 테스트 실행 (현재는 실패)
cd 01-flat/exercise
go test -v

# 3. flat.go 구현 시작
# - NewFlatIndex()
# - Add()
# - Search()
# - Size()

# 4. 테스트 통과 확인
go test -v

# 5. Solution과 비교
cd ../solution
cat EXPLANATION.md
```

## 📚 학습 경로

### Week 1: Flat Index 정복
- [x] pkg/ 유틸리티 이해
- [ ] 01-flat/README.md 읽기
- [ ] 01-flat/exercise/flat.go 구현
- [ ] 모든 테스트 통과
- [ ] Benchmark 실행 및 결과 분석
- [ ] Solution과 비교

**목표**: 벡터 검색의 기본 이해, O(n) baseline 확립

### Week 2-3: IVF Index로 가속화
- [ ] 02-ivf/README.md 읽기
- [ ] k-means 알고리즘 구현
- [ ] IVF 인덱스 구현
- [ ] **함정 경험**: nprobe=1일 때 recall 저하
- [ ] 파라미터 튜닝으로 recall 향상
- [ ] Flat vs IVF 성능 비교

**목표**: 클러스터링 개념, 속도/정확도 트레이드오프 체득

### Week 4-6: HNSW Index 마스터
- [ ] 03-hnsw/README.md 읽기
- [ ] 그래프 구조 설계
- [ ] 계층적 삽입 알고리즘 구현
- [ ] 탐색 알고리즘 구현
- [ ] M, efConstruction, efSearch 파라미터 튜닝
- [ ] 세 인덱스 종합 비교

**목표**: 최신 ANN 기술 이해, 복잡한 파라미터 최적화

### Week 7: 종합 및 심화
- [ ] examples/ 실행하여 실전 활용
- [ ] docs/THEORY.md로 이론 심화
- [ ] 자신만의 최적화 시도
- [ ] 실제 데이터로 실험

## 🧪 테스트 전략

### 기본 테스트
```bash
cd 01-flat/exercise
go test -v                    # 모든 테스트
go test -v -run=TestBasicAdd  # 특정 테스트
```

### Race Condition 검사
```bash
go test -race                 # 동시성 문제 탐지
```

### 벤치마크
```bash
go test -bench=.              # 성능 측정
go test -bench=. -benchmem    # 메모리 할당 포함
```

## 📊 예상 성능 (참고용)

| Index | Build Time | Search Time (k=10) | Recall | Memory |
|-------|-----------|-------------------|--------|--------|
| Flat  | O(1) | O(n) | 100% | O(n·d) |
| IVF   | O(training) | O(nprobe·cluster) | 85-95% | O(n·d) |
| HNSW  | O(n log n) | O(log n) | 95-99% | O(n·d·M) |

*n = 벡터 수, d = 차원, cluster = 클러스터당 벡터 수*

## 💡 학습 팁

### 1. 테스트 주도 개발 (TDD)
```bash
# Red: 실패하는 테스트 확인
go test -v -run=TestBasicAdd

# Green: 최소한으로 통과시키기
# 구현...

# Refactor: 리팩토링
# 개선...
```

### 2. 에러 메시지를 친구로
```
=== RUN   TestDimensionMismatch
    flat_test.go:45: Add() should fail with dimension mismatch
--- FAIL: TestDimensionMismatch (0.00s)
```
→ 차원 검증 로직 추가 필요!

### 3. Solution은 막혔을 때만
1. 먼저 스스로 구현 시도 (30분~1시간)
2. 힌트 다시 읽기
3. 관련 pkg/ 코드 참고
4. 그래도 안 되면 solution/ 확인
5. 이해한 후 자신의 코드로 다시 작성

### 4. 벤치마크로 학습
```bash
# 벡터 수 증가 시 시간 변화 관찰
BenchmarkSearch/size=100    →  250µs
BenchmarkSearch/size=1000   → 2500µs (10배)
BenchmarkSearch/size=10000  → 25000µs (10배)
```
→ O(n) 복잡도 확인!

## 🎓 학습 목표

이 코스를 완료하면:

- ✅ 벡터 유사도 검색의 기본 원리 이해
- ✅ 브루트포스, 클러스터링, 그래프 기반 알고리즘 구현 능력
- ✅ ANN (Approximate Nearest Neighbors) 개념 체득
- ✅ 파라미터 튜닝 경험 (nprobe, M, efSearch)
- ✅ 속도-정확도-메모리 트레이드오프 이해
- ✅ Go 동시성, 벤치마킹, 테스팅 실력 향상
- ✅ Faiss, Milvus, Weaviate 등 실제 벡터 DB 이해력 향상

## 🔗 추가 자료

### 논문
- [Efficient and robust approximate nearest neighbor search using Hierarchical Navigable Small World graphs](https://arxiv.org/abs/1603.09320) (HNSW)
- [Product Quantization for Nearest Neighbor Search](https://hal.inria.fr/inria-00514462v2/document) (PQ, IVF)

### 오픈소스 라이브러리
- [Faiss](https://github.com/facebookresearch/faiss) - Facebook의 벡터 검색 라이브러리
- [hnswlib](https://github.com/nmslib/hnswlib) - HNSW C++ 구현
- [Annoy](https://github.com/spotify/annoy) - Spotify의 ANN 라이브러리

### 벡터 데이터베이스
- [Milvus](https://milvus.io/)
- [Weaviate](https://weaviate.io/)
- [Qdrant](https://qdrant.tech/)

## 🤝 기여

이 과정을 개선할 아이디어가 있나요?

1. Issue 생성
2. 개선 사항 제안
3. Pull Request 제출

## 📄 라이선스

MIT License

## 🙋 FAQ

**Q: Go를 잘 모르는데 괜찮나요?**
A: 기본 문법 (함수, 구조체, 슬라이스, 에러 처리)만 알면 됩니다. 코드를 보며 배울 수 있습니다.

**Q: 수학이 어려운데요?**
A: 벡터 거리 계산 정도만 이해하면 됩니다. pkg/distance/가 이미 구현되어 있습니다.

**Q: 얼마나 시간이 걸리나요?**
A: 주당 5-10시간 투자 시 6-7주 완료 가능합니다. 자신의 속도로 진행하세요!

**Q: 실무에서 쓸 수 있나요?**
A: 이 구현은 학습용입니다. 실무에서는 Faiss, Milvus 등을 사용하세요. 하지만 이 과정 후 그 라이브러리들을 훨씬 잘 이해할 수 있습니다!

**Q: Flat만 하고 다음에 IVF 해도 되나요?**
A: 네! 각 인덱스는 독립적입니다. 하지만 순서대로 하는 것을 권장합니다.

## 🎉 시작하기

준비되셨나요? 첫 걸음을 시작하세요:

```bash
cd 01-flat
cat README.md
```

Good luck, and happy coding! 🚀

---

*이 프로젝트는 벡터 데이터베이스와 ANN 알고리즘을 깊이 이해하고자 하는 개발자들을 위해 만들어졌습니다.*
