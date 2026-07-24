# 테넌트 Namespace 삭제

## 사전 확인
1. 해당 namespace(예: enterprise)의 모든 워크로드 식별 — `kubectl get all -n <ns>`
2. PVC·Secret·ConfigMap 등 영구 자원 백업 필요성 판단
3. cross-namespace 참조 확인 — enterprise는 notiflex의 Valkey(`valkey-primary.notiflex.svc.cluster.local`)를 공유한다. 삭제해도 notiflex 쪽 Valkey에는 영향 없음(참조 방향이 단방향)
4. ArgoCD Application(`notiflex-<tenant>`)이 이 namespace를 관리하는지 확인 — `kubectl get app -n argocd`

## 실행
1. root-app의 `include` 목록에서 해당 테넌트를 제거하고 `k8s/<tenant>/`를 삭제 → git push
   - 부모 App of Apps는 자기 미관리이므로 `kubectl apply -f k8s/root-app.yaml`로 include 갱신 반영
2. ArgoCD가 자식 Application과 그 안의 리소스를 prune하기를 대기
3. 잔여 리소스가 있으면 매니페스트에서 제거하고 git push (kubectl delete 직접 사용 금지 — selfHeal이 되돌림)
4. GSM IAM 바인딩 정리 — 해당 테넌트 KSA를 `terraform/gcp/apps`의 `gsm_secret_accessors` 맵에서 제거 후 apply

## 사후 검증
1. ArgoCD UI에서 Application이 사라졌는지 확인
2. `kubectl get all -n <namespace>`로 리소스가 모두 정리됐는지 확인
3. namespace 자체 삭제 여부 확인 (`kubectl get ns`)
