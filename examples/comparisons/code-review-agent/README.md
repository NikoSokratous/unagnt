# Code Review Agent Comparison

This example implements the **same complex code review workflow** using different frameworks to demonstrate the differences in approach, complexity, and capabilities.

## The Use Case

A production-grade automated code review agent that:
1. Fetches code changes from a GitHub PR
2. Performs static analysis (linting, security scanning)
3. Reviews code quality and suggests improvements
4. Checks for breaking changes
5. Generates a comprehensive review report
6. Posts comments back to the PR
7. Implements retry logic and error handling
8. Tracks costs and enforces policies

## Implementations

### 1. Unagnt (`./unagnt/`)
**Lines of Code**: ~150 (YAML + minimal glue)  
**Features**: All capabilities built-in

### 2. LangGraph (`./langgraph/`)
**Lines of Code**: ~450 (Python)  
**Features**: Requires extensive custom code

## Key Differences

| Feature | Unagnt | LangGraph |
|---------|-------------|-----------|
| **Policy Enforcement** | ✅ Built-in YAML policies | ❌ Must implement manually |
| **Cost Tracking** | ✅ Automatic per-agent/tenant | ❌ Must implement manually |
| **Observability** | ✅ Built-in tracing/metrics | ⚠️ Requires LangSmith (paid) |
| **Human-in-the-Loop** | ✅ Declarative approval gates | ⚠️ Manual implementation |
| **Retry Logic** | ✅ Built-in with backoff | ❌ Must implement manually |
| **Multi-tenancy** | ✅ Built-in namespace isolation | ❌ Must implement manually |
| **Deterministic Replay** | ✅ 5 replay modes built-in | ❌ Not available |
| **Security** | ✅ Permission-gated tools | ❌ Must implement manually |
| **Production Ready** | ✅ Day 1 | ⚠️ Requires significant dev work |

## Running the Examples

### Unagnt
```bash
cd unagnt
unagnt workflow run code-review.yaml \
  --param pr_url=https://github.com/user/repo/pull/123 \
  --param severity_threshold=medium
```

### LangGraph
```bash
cd langgraph
pip install -r requirements.txt
python code_review_agent.py --pr-url https://github.com/user/repo/pull/123
```

## Complexity Comparison

### Configuration vs Code

**Unagnt**: Declarative YAML configuration
- Workflow: 80 lines of YAML
- Policy: 40 lines of YAML
- Glue code: 30 lines of Go (custom tools if needed)
- **Total**: ~150 lines

**LangGraph**: Imperative Python code
- Main agent: 200 lines
- State management: 80 lines
- Error handling: 70 lines
- Observability: 50 lines
- Cost tracking: 50 lines
- **Total**: ~450 lines

### What You Get Out-of-the-Box

#### Unagnt
```yaml
# Just describe what you want
name: "code-review-agent"
policy: "production-safety"  # Built-in governance
cost_tracking: true          # Automatic
human_approval:              # Declarative
  required_for: ["post_comment"]
retry:                       # Built-in
  max_attempts: 3
  backoff: exponential
```

#### LangGraph
```python
# Must implement everything yourself
class CodeReviewAgent:
    def __init__(self):
        self.setup_state()           # Manual
        self.setup_retry_logic()     # Manual
        self.setup_cost_tracking()   # Manual
        self.setup_error_handling()  # Manual
        self.setup_approval_gates()  # Manual
        # 200+ more lines...
```

## Production Considerations

### Unagnt Advantages
1. **Policy Enforcement**: YAML-based governance from day 1
2. **Observability**: Built-in tracing, metrics, replay
3. **Security**: Permission system prevents dangerous operations
4. **Cost Control**: Automatic tracking with budget alerts
5. **Multi-tenancy**: Namespace isolation built-in
6. **Debugging**: Deterministic replay of any execution
7. **Scalability**: Kubernetes-native with auto-scaling

### LangGraph Advantages
1. **Flexibility**: Full Python control for custom logic
2. **Ecosystem**: Rich library of pre-built components
3. **Research**: Good for experimentation and prototyping
4. **Community**: Large community and examples

## When to Use What?

### Choose Unagnt if:
- ✅ You need production-grade features immediately
- ✅ You want declarative, maintainable workflows
- ✅ Governance and compliance are important
- ✅ You need cost tracking and control
- ✅ You're building for enterprise/production
- ✅ You want Kubernetes-native deployment

### Choose LangGraph if:
- ✅ You're doing research or prototyping
- ✅ You need highly custom logic in every step
- ✅ You're comfortable implementing production features yourself
- ✅ You're already invested in the LangChain ecosystem

## Real-World Impact

### Development Time
- **Unagnt**: 2-3 hours to production-ready workflow
- **LangGraph**: 2-3 days to add production features

### Maintenance
- **Unagnt**: Update YAML config, policies versioned
- **LangGraph**: Code changes, testing, deployment

### Debugging
- **Unagnt**: Replay any execution deterministically
- **LangGraph**: Add logging, hope you captured enough

### Scaling
- **Unagnt**: `kubectl scale` or HPA
- **LangGraph**: Custom infrastructure, load balancing, etc.

## Conclusion

Both frameworks have their place:
- **LangGraph**: Great for prototyping and custom research
- **Unagnt**: Built for production from day one

If you're building a proof-of-concept, either works. If you're building production AI systems, Unagnt gives you the infrastructure you need without writing it yourself.

## Try It Yourself

Run both implementations and compare:
1. Ease of setup
2. Code complexity
3. Production features available
4. Debugging experience
5. Deployment story

The code speaks for itself! 🚀
