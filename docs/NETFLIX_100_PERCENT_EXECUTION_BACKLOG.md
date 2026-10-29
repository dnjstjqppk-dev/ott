# NETFLIX 100% PLATFORM EXECUTION BACKLOG

본 문서는 현재 저장소 상태를 기준으로, Netflix와 동등한 글로벌 OTT 플랫폼이 되기 위해 **남은 모든 작업**을 Phase별로 정리한 실행 백로그이다.

---

## 0. 현재 상태 요약

### 완료/부분 완료
- Web User: 홈/행/히어로/상세 모달/인증 모달/검색/찜하기/시청중 행/재생 분기
- Web User 재생: Shaka Player 구조, DRM 헤더, Trickplay, Heartbeat 구조
- Backend 서비스 코드: API Gateway, user-profile, catalog, playback, billing, reco, media-worker, studio-ingest, open-connect-edge
- Infra/매니페스트: K8s, Cloudflare, wrangler, CI gate
- 문서: 10 Phase 마스터 플랜

### 미완료/미검증
- 서비스 실제 기동 및 API 연동
- DRM 라이선스 실교환
- ABR/VMAF/인코딩 파이프라인 실검증
- 추천/MAB 피드백 루프 실검증
- 데이터 플랫폼 대규모 처리 검증
- SRE/Chaos/Failover 실검증
- 멀티플랫폼 클라이언트 완성

---

## 1. Phase 1 — 글로벌 엣지 & Open Connect CDN

### 남은 작업
- [ ] BGP Anycast 데몬 연동 및 ISP ASN 피어링 검증
- [ ] GeoDNS/Proximity 라우팅 알고리즘 구현 및 RTT 기반 최적 노드 산출
- [ ] OCA 장애 감지 및 50ms 이내 failover 검증
- [ ] Signed URL / Edge Token HMAC 검증 end-to-end 테스트
- [ ] 캐시 적중률 95% 이상 측정 파이프라인 구축
- [ ] 글로벌 지연 < 15ms 측정 및 모니터링

### 완료 기준
- 글로벌 지연 < 15ms
- 캐시 적중률 > 95%

---

## 2. Phase 2 — 미디어 인코딩 & DRM

### 남은 작업
- [ ] Shot-based Dynamic Optimizer 구현
- [ ] VMAF 4K 품질 측정 엔진 연동
- [ ] CMAF 패키징 파이프라인 구축
- [ ] Widevine L1 / FairPlay / PlayReady 라이선스 서버 연동
- [ ] DRM 세션 ID / 토큰 교환 end-to-end 검증
- [ ] 적응형 비트레이트 스트리밍 실제 재생 검증
- [ ] Trickplay sprite 생성 및 VTT 파싱 검증
- [ ] 광고 삽입 cue point 처리 검증

### 완료 기준
- VMAF >= 95 기준 비트레이트 40% 절감
- 3대 DRM 실제 연동

---

## 3. Phase 3 — 코어 백엔드 & 결제

### 남은 작업
- [ ] API Gateway 라우팅/인증/Rate Limit 실검증
- [ ] IAM / Household Verification 실구현 검증
- [ ] 프로필 PIN 검증 API 연동 검증
- [ ] Catalog API 실서버 연동 검증
- [ ] 구독 결제 원장(Billing) 실구현 검증
- [ ] 결제 실패 스마트 재시도 구현
- [ ] 초당 50,000 TPS 부하 테스트
- [ ] DB 마이그레이션 및 트랜잭션 정합성 검증

### 완료 기준
- 초당 50,000 TPS 처리
- 결제 실패 스마트 재시도

---

## 4. Phase 4 — AI 개인화 & MAB

### 남은 작업
- [ ] MAB 썸네일 개인화 엔진 실구현
- [ ] Two-Tower 신경망 추천 엔진 실구현
- [ ] XP A/B 테스팅 플랫폼 구현
- [ ] 프로필별 맞춤 포스터 노출 검증
- [ ] 클릭 이벤트 수집 피드백 루프 활성화
- [ ] 추천 CTR +25% 측정

### 완료 기준
- 추천 클릭률(CTR) +25%
- A/B 통계 검정 자동화

---

## 5. Phase 5 — 스트리밍 & ABR

### 남은 작업
- [ ] BBA/MPC 적응형 알고리즘 실구현
- [ ] 실시간 재생 텔레메트리/QoS 구현
- [ ] 북마크 싱크 P99 < 100ms 검증
- [ ] 세션 동시 시청 제어 구현
- [ ] 재버퍼링 비율(VBR) < 0.25% 검증
- [ ] 광고 삽입/스킵 마커 실동작 검증

### 완료 기준
- 재버퍼링 비율(VBR) < 0.25%
- 북마크 싱크 P99 < 100ms

---

## 6. Phase 6 — 멀티플랫폼 클라이언트

### Web User 남은 작업
- [ ] 검색 고도화(자동완성, 필터, 정렬)
- [ ] My List 영속화/동기화
- [ ] Hover Preview 구현
- [ ] Continue Watching 정확한 북마크 반영
- [ ] 프로필 관리/전환 UX 강화
- [ ] 알림/공지 UI
- [ ] 오프라인 다운로드
- [ ] 접근성(a11y) 및 키보드 내비게이션
- [ ] 성능 최적화(이미지/번들)

### Web Admin / Studio
- [ ] Admin 대시보드 실데이터 연동
- [ ] Studio 파트너 업로드/인코딩 상태 연동

### Mobile / TV
- [ ] Flutter/Swift/Kotlin 클라이언트 구현
- [ ] Tizen/webOS/Android TV 10-foot UI 구현

### 완료 기준
- 전 플랫폼 60fps 렌더링
- 오프라인 암호화 다운로드

---

## 7. Phase 7 — 데이터 플랫폼

### 남은 작업
- [ ] Kafka 토픽 파티셔닝 및 Schema Registry 검증
- [ ] Flink 실시간 트렌딩 집계 구현
- [ ] 계정 탈취/스크래핑 탐지 구현
- [ ] Iceberg/ClickHouse 레이크하우스 구축
- [ ] 일 1,000억 건 이벤트 처리 검증
- [ ] 무손실 실시간 랭킹 산출 검증

### 완료 기준
- 일 1,000억 건 이벤트 처리
- 무손실 실시간 랭킹 산출

---

## 8. Phase 8 — 보안 & 거버넌스

### 남은 작업
- [ ] Cloud KMS/HSM 키 계층 구현
- [ ] DRM 라이선스 토큰 검증
- [ ] 화면 캡처 방지 검증
- [ ] GDPR/CCPA/ISMS-P 준수 자동화
- [ ] Right to be Forgotten 자동화
- [ ] PCI-DSS Level 1 결제 데이터 토큰화

### 완료 기준
- 하드웨어 레벨 스크린 캡처 차단
- 결제 데이터 격리

---

## 9. Phase 9 — SRE & Chaos

### 남은 작업
- [ ] Multi-Region Active-Active 구축
- [ ] 자동 페일오버 검증
- [ ] Chaos Monkey/Chaos Mesh 실험
- [ ] Circuit Breaker 검증
- [ ] Canary 배포 파이프라인 구현
- [ ] 자동 롤백 구현
- [ ] 99.999% 가용성 측정

### 완료 기준
- 리전 장애 시 무중단 자동 페일오버
- 가용성 99.999%

---

## 10. Phase 10 — 라이브 & 인터랙티브

### 남은 작업
- [ ] LL-CMAF/LL-HLS 초저지연 라이브 스트리밍
- [ ] 라이브 채팅/리액션 브로드캐스트
- [ ] Branching Narrative Engine 구현
- [ ] Seamless Splice 씬 전환 구현
- [ ] 선택 상태 그래프 트리 저장

### 완료 기준
- 5,000만 동시 라이브 접속 지연 없음
- 무봉제 씬 전환

---

## 11. 공통/품질 작업

### 테스트
- [ ] Web User E2E 테스트 (Playwright/Cypress)
- [ ] Backend 단위/통합 테스트 보강
- [ ] DRM/스트리밍 E2E 테스트
- [ ] 부하/카오스 테스트 자동화

### CI/CD
- [ ] 빌드/테스트/린트 게이트 강화
- [ ] 마이그레이션 자동화
- [ ] 배포 파이프라인 검증

### 관측성
- [ ] 메트릭/로그/트레이싱 통합
- [ ] SLO/SLI 정의
- [ ] 알림/온콜 체계

---

## 12. 우선 실행 순서 (권장)

1. **Phase 3 백엔드 실연동**
2. **Phase 2 DRM/스트리밍 실연동**
3. **Phase 4 추천 피드백 루프**
4. **Phase 6 Web User 고급 UX**
5. **Phase 7/8/9/10 인프라 및 고급 기능**

---

## 13. 현재 바로 해야 할 Next Actions

- [ ] API Gateway + Web User 로컬 연동 테스트
- [ ] Playback Session Ticket 실제 발급/재생 검증
- [ ] Auth/Profile/PIN E2E 검증
- [ ] Search/MyList/Continue Watching UX 검증
- [ ] Shaka Player 실스트림 재생 검증
- [ ] CI build/test gate 실행
- [ ] 백엔드 서비스별 README/실행 가이드 정리
- [ ] 통합 테스트 시나리오 문서화
