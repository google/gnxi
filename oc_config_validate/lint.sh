#!/bin/bash
# Lint Python and YAML files of the project

BASEDIR=$(dirname $0)

echo "---Linting JSON files in ${BASEDIR}/init_configs"
for f in "${BASEDIR}"/init_configs/*.json; do
  python3 -m json.tool "$f" /dev/null || echo "Errors in $f"
done

echo "---Linting YAML files in ${BASEDIR}/tests"
yamllint -d relaxed "${BASEDIR}"/tests/

echo "---Formatting and Linting Python files in ${BASEDIR}/oc_config_validate"
python3 -m autopep8 -i "${BASEDIR}"/oc_config_validate/*.py "${BASEDIR}"/oc_config_validate/testcases/*.py
python3 -m isort "${BASEDIR}"/oc_config_validate/*.py "${BASEDIR}"/oc_config_validate/testcases/*.py
python3 -m pylama -l pycodestyle,pyflakes "${BASEDIR}"/oc_config_validate/*.py "${BASEDIR}"/oc_config_validate/testcases/*.py
python3 -m pytype "${BASEDIR}"/oc_config_validate/*.py "${BASEDIR}"/oc_config_validate/testcases/*.py -x "${BASEDIR}"/oc_config_validate/models
