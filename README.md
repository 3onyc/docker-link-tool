# docker-link-tool

Tool for linking docker containers without using explicit links, primarily useful with --icc=false

## Examples

### Add rule allowing container foo to access port 80 on container bar

```bash
link-tool -action add -cname foo -sname bar -port 80
```

### Remove rule allowing foo to access port 80 on bar

```bash
link-tool -action delete -cname foo -sname bar -port 80
```
