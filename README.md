# tampl - dead simple config file templating

`tampl` is a small application that renders configuration (and other) files by rendering templates.
It uses standard [go templates](https://golang.org/pkg/text/template/) and a [YAML](http://yaml.org/) file to define template variables.

To keep it simple only two arguments are expected, first the source directory containing the `_vars.yml` and a number of template files ending in `.tmpl`, and second the target directory's path where the rendered templates will be stored.
The target files will have the same name as the equivalent template file but with the `.tmpl` suffix removed.

```
$ tampl
tampl <source> <target>
	source: a directory containing a number of '.tmpl' files and '_vars.yml' which stores the variables and values.
	target: the directory where the rendered templates should be placed into.
```