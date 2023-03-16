# panther

`panther` is an AI cat who is good at software engineering.
It uses ChatGPT (GPT-3.5) to explain codes in PRs and programs. Currently `panther` supports:

- `explain-pr`
- `summarize-pr`
- `explain-fn`


## Explain a PR

```
# Explain codes in the PR #500 of the TF-controller repository.

panther explain-pr -r weaveworks/tf-controller 500

# Explain codes in the PR #4000 of Weave GitOps repository.

panther explain-pr 4000
```

## Summarize a PR

This function uses the result from `explain-pr` command to summarize the changes made for a specific PR.

```
# Summarize code explanation in the PR #4000 of Weave GitOps repository.

panther summarize-pr 4000
```


## Explain a function

It supports only Go function at the moment.

```
# Explain codes in the function named "runExplainPRCmd" in a local repo.

panther explain-fn "runExplainPRCmd"
```
