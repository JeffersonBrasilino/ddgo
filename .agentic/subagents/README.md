# `.agentic/subagents/` — Instruções canônicas de subagentes

Esta pasta guarda **somente as instruções canônicas** de cada subagent — o
prompt que define papel, metodologia, formato de saída e restrições. As
instruções são **agnósticas de harness e de modelo/LLM**: nada aqui contém
particularidades como `mode`, `permissions`, `tools` ou `model`.

As particularidades de cada harness (o frontmatter que ele espera) ficam em
**arquivos próprios no diretório de descoberta de cada harness**, cujo corpo é
apenas uma referência às instruções canônicas desta pasta — exatamente o mesmo
padrão usado em `.opencode/skills/`:

```
.agentic/subagents/CodeReviewer.md      ← instruções reais (fonte única)
.opencode/agents/CodeReviewer.md        ← frontmatter OpenCode V2 + "read `.agentic/...`"
.claude/agents/CodeReviewer.md          ← frontmatter Claude Code  + "read `.agentic/...`"
```

## CodeReviewer

**Papel**: Tech Lead sênior de Go (design patterns, SOLID, arquitetura,
concorrência, DDD) que faz code review **educacional e read-only**.

- Não corrige nada — apenas devolve apontamentos de melhoria e melhores
  abordagens, cada um com **motivador detalhado** (consequência prática,
  princípio por trás e cenário concreto de falha);
- Revisa apenas o que fizer sentido: prioriza escopo explícito do usuário,
  API pública, núcleo de domínio, caminhos de erro/concorrência e testes;
- Usa as skills samber de `.agentic/skills/` como rubrica de análise
  (exceto `adjust-go-code`, que modifica código — proibida para ele);
- Saída: relatório com escopo, sumário executivo, pontos fortes,
  apontamentos ordenados por severidade e trilha de aprendizado.

## Como conectar um subagent a um novo harness

Repita o padrão acima para qualquer harness/LLM:

1. Crie o arquivo no **diretório de descoberta de agentes** do harness
   (ex.: `.opencode/agents/`, `.claude/agents/`, ou o equivalente);
2. No **frontmatter**, coloque apenas as particularidades que esse harness
   espera (descrição, modo, permissões/ferramentas, restrições);
3. No **corpo**, referencie as instruções canônicas:

   ```markdown
   read `.agentic/subagents/CodeReviewer.md`.
   ```

Para uma LLM/harness sem mecanismo de descoberta (uso direto via API), use o
conteúdo de `.agentic/subagents/CodeReviewer.md` como *system prompt* — ele é
autocontido e funciona sozinho.

### OpenCode (V2) — já conectado

`.opencode/agents/CodeReviewer.md` com `mode: all` (usável como subagent
delegado ou como agente primário da sessão) e `permissions` que bloqueiam
`edit`, `subagent` e os comandos de shell que mutam o workspace — garantindo
o read-only em nível de ferramenta, não só de prompt. Uso:

```text
Use o subagent CodeReviewer para revisar o pacote domain e me trazer
apontamentos de melhoria.
```

### Claude Code — já conectado

`.claude/agents/CodeReviewer.md` com `tools` restrito a leitura e análise
(`Read`, `Grep`, `Glob`, `Bash`). O agente principal delega via Task/subagent
quando você pedir um code review.

## Agnosticismo

- **Sem `model` em nenhum arquivo**: o subagent herda o modelo da sessão em
  qualquer harness — trocar de LLM não exige mudança alguma;
- **Read-only em camadas**: permissões/ferramentas restritas no arquivo do
  harness (onde suportado) + "regra de ouro" nas instruções canônicas +
  exclusão da skill `adjust-go-code` (a única skill samber que modifica
  código);
- **Fonte única de verdade**: instruções editadas uma vez em
  `.agentic/subagents/CodeReviewer.md` valem para todos os harnesses;
  particularidades de cada harness vivem isoladas no arquivo dele.
