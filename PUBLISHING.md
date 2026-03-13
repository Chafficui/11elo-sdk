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

Register the package on [packagist.org](https://packagist.org/) pointing to this repository's `php/` directory. New versions are picked up automatically via git tags.

## Release checklist

1. Update version numbers in all package manifests
2. Update CHANGELOGs (if applicable)
3. Ensure all tests pass (`CI` workflow on GitHub)
4. Tag the release: `git tag v0.2.0 && git push origin v0.2.0`
5. Publish each package (see commands above)
