# Publishing

Internal reference for releasing new SDK versions.

## JavaScript (npm)

```bash
cd javascript
# bump version in package.json
npm publish
```

## Python (PyPI)

```bash
cd python
# bump version in pyproject.toml
python -m build
twine upload dist/*
```

## Go

Go modules are published automatically when you tag a release:

```bash
# bump version tag
git tag go/v0.2.0
git push origin go/v0.2.0
```

## PHP (Packagist)

The `php/` directory is automatically mirrored to [github.com/Chafficui/elevenelo-php](https://github.com/Chafficui/elevenelo-php) via a GitHub Actions subtree split (`.github/workflows/split-php.yml`). Packagist is pointed at that mirror repo.

**Setup (one-time):**
1. Add a PAT with `repo` scope as the `SUBTREE_PAT` secret in the 11elo-sdk repo settings
2. Register `Chafficui/elevenelo-php` on [packagist.org](https://packagist.org/)

New versions are picked up automatically via git tags on the mirror repo.

## Release checklist

1. Update version numbers in all package manifests
2. Update CHANGELOGs (if applicable)
3. Ensure all tests pass (`CI` workflow on GitHub)
4. Tag the release: `git tag v0.2.0 && git push origin v0.2.0`
5. Publish each package (see commands above)
