---
# Particularidades do OpenCode V2 para o subagent CodeReviewer.
# As instruções reais (canônicas) vivem em .agentic/subagents/CodeReviewer.md —
# agnósticas de harness e modelo. Nenhuma chave `model` é definida: o subagent
# herda o modelo da sessão.
description: >-
  Tech Lead sênior de Go que realiza code review educacional read-only: analisa
  o que fizer sentido no codebase e devolve apontamentos de melhoria e melhores
  abordagens com motivadores detalhados (design, SOLID, arquitetura,
  idiomatismo Go, concorrência, error handling, testes) — sem corrigir nada.
  Use quando pedirem review, feedback técnico, apontamentos de melhorias ou
  melhores abordagens em código Go.
mode: all
permissions:
  # Este subagent é SOMENTE LEITURA: bloqueia qualquer modificação de arquivos.
  - action: edit
    resource: "*"
    effect: deny
  # Defense-in-depth: bloqueia os comandos de shell mais comuns que mutam o workspace.
  - action: shell
    resource: "gofmt -w*"
    effect: deny
  - action: shell
    resource: "goimports -w*"
    effect: deny
  - action: shell
    resource: "go mod tidy*"
    effect: deny
  - action: shell
    resource: "go mod edit*"
    effect: deny
  - action: shell
    resource: "go generate*"
    effect: deny
  - action: shell
    resource: "git add*"
    effect: deny
  - action: shell
    resource: "git commit*"
    effect: deny
  - action: shell
    resource: "git push*"
    effect: deny
  - action: shell
    resource: "git checkout*"
    effect: deny
  - action: shell
    resource: "git restore*"
    effect: deny
  - action: shell
    resource: "git stash*"
    effect: deny
  # Um reviewer não precisa delegar trabalho para outros agentes.
  - action: subagent
    resource: "*"
    effect: deny
---

read `.agentic/subagents/code_reviewer.md`.
