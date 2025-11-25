#!/bin/bash
# Popular Packages Database
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Database of popular packages for typosquatting detection and security analysis.
# Used to identify potentially malicious packages with names similar to popular ones.

set -eo pipefail

#############################################################################
# Popular Package Lists by Ecosystem
# These lists are used for:
# 1. Typosquatting detection (similar names to popular packages)
# 2. Library recommendations (known replacements)
# 3. Supply chain risk assessment
#############################################################################

# Store popular packages as newline-separated strings for compatibility
NPM_POPULAR_LIST="lodash
underscore
ramda
express
koa
fastify
hapi
restify
react
vue
angular
svelte
preact
next
nuxt
gatsby
redux
mobx
zustand
recoil
jotai
axios
request
got
node-fetch
ky
superagent
moment
dayjs
date-fns
luxon
chalk
colors
debug
commander
yargs
async
bluebird
rxjs
uuid
nanoid
shortid
jest
mocha
chai
jasmine
ava
cypress
playwright
puppeteer
sinon
nock
webpack
rollup
parcel
esbuild
vite
babel
typescript
swc
eslint
prettier
husky
mongoose
sequelize
knex
prisma
typeorm
mysql
mysql2
pg
mongodb
redis
ioredis
passport
jsonwebtoken
bcrypt
bcryptjs
argon2
joi
yup
zod
validator
class-validator
fs-extra
glob
rimraf
mkdirp
chokidar
multer
formidable
winston
pino
bunyan
morgan
log4js
dotenv
config
convict
pm2
nodemon
forever
ws
crypto-js
node-forge
aws-sdk
openai
anthropic
langchain
cohere-ai
prop-types
classnames
clsx
body-parser
cors
helmet
sharp
jimp
cheerio
jsdom
handlebars
ejs
pug
marked
markdown-it
semver
minimist
meow
ora
inquirer
prompts
cross-env
concurrently
npm-run-all
lerna
nx
turbo"

PYPI_POPULAR_LIST="django
flask
fastapi
tornado
bottle
starlette
sanic
aiohttp
requests
httpx
urllib3
numpy
pandas
scipy
matplotlib
seaborn
scikit-learn
sklearn
statsmodels
tensorflow
torch
pytorch
keras
transformers
huggingface-hub
openai
anthropic
langchain
llama-index
cohere
sqlalchemy
psycopg2
pymongo
redis
alembic
peewee
tortoise-orm
boto3
botocore
awscli
s3transfer
pytest
unittest2
nose
mock
pytest-cov
coverage
tox
nox
hypothesis
faker
factory-boy
click
typer
argparse
pyyaml
toml
python-dotenv
python-dateutil
pytz
arrow
pendulum
pydantic
attrs
dataclasses
typing-extensions
mypy
pylint
flake8
black
isort
autopep8
cryptography
pycryptodome
passlib
pyjwt
python-jose
pillow
opencv-python
imageio
asyncio
aiofiles
uvloop
beautifulsoup4
bs4
scrapy
selenium
lxml
djangorestframework
marshmallow
graphene
celery
rq
dramatiq
loguru
structlog
google-cloud-storage
azure-storage-blob
six
setuptools
wheel
pip
jinja2
werkzeug
markupsafe
chardet
charset-normalizer
certifi
idna
colorama
tqdm
rich
regex
more-itertools
toolz
joblib
multiprocess"

GO_POPULAR_LIST="github.com/gin-gonic/gin
github.com/labstack/echo
github.com/gofiber/fiber
github.com/gorilla/mux
github.com/go-chi/chi
google.golang.org/grpc
google.golang.org/protobuf
gorm.io/gorm
github.com/jmoiron/sqlx
github.com/go-redis/redis
go.mongodb.org/mongo-driver
github.com/aws/aws-sdk-go
github.com/aws/aws-sdk-go-v2
go.uber.org/zap
github.com/sirupsen/logrus
github.com/stretchr/testify
github.com/onsi/ginkgo
github.com/onsi/gomega
github.com/spf13/cobra
github.com/spf13/viper
github.com/urfave/cli
github.com/google/uuid
github.com/pkg/errors
golang.org/x/crypto
golang.org/x/net"

MAVEN_POPULAR_LIST="org.springframework:spring-core
org.springframework:spring-web
org.springframework.boot:spring-boot
org.springframework.boot:spring-boot-starter
org.slf4j:slf4j-api
ch.qos.logback:logback-classic
org.apache.logging.log4j:log4j-core
com.fasterxml.jackson.core:jackson-databind
com.google.code.gson:gson
junit:junit
org.junit.jupiter:junit-jupiter
org.mockito:mockito-core
org.apache.httpcomponents:httpclient
com.squareup.okhttp3:okhttp
org.hibernate:hibernate-core
mysql:mysql-connector-java
org.postgresql:postgresql
org.apache.commons:commons-lang3
com.google.guava:guava
org.projectlombok:lombok"

# Known malicious/compromised packages
NPM_MALICIOUS_LIST="event-stream
flatmap-stream
colors
faker
ua-parser-js
coa
rc
left-pad"

PYPI_MALICIOUS_LIST="python-dateutil
jeIlyfish"

#############################################################################
# Package Lookup Functions
#############################################################################

# Check if package is in popular list
# Usage: is_popular_package <package> <ecosystem>
is_popular_package() {
    local pkg="$1"
    local ecosystem="${2:-npm}"
    local list=""

    case "$ecosystem" in
        npm|node)
            list="$NPM_POPULAR_LIST"
            ;;
        pypi|python)
            list="$PYPI_POPULAR_LIST"
            ;;
        go|golang)
            list="$GO_POPULAR_LIST"
            ;;
        maven|java)
            list="$MAVEN_POPULAR_LIST"
            ;;
        *)
            echo "false"
            return
            ;;
    esac

    # Use grep to check if package is in list (exact match)
    if echo "$list" | grep -qx "$pkg"; then
        echo "true"
    else
        echo "false"
    fi
}

# Get list of popular packages for ecosystem
# Usage: get_popular_packages <ecosystem>
get_popular_packages() {
    local ecosystem="${1:-npm}"

    case "$ecosystem" in
        npm|node)
            echo "$NPM_POPULAR_LIST"
            ;;
        pypi|python)
            echo "$PYPI_POPULAR_LIST"
            ;;
        go|golang)
            echo "$GO_POPULAR_LIST"
            ;;
        maven|java)
            echo "$MAVEN_POPULAR_LIST"
            ;;
    esac
}

# Get count of popular packages for ecosystem
# Usage: get_popular_package_count <ecosystem>
get_popular_package_count() {
    local ecosystem="${1:-npm}"
    local list=""

    case "$ecosystem" in
        npm|node)
            list="$NPM_POPULAR_LIST"
            ;;
        pypi|python)
            list="$PYPI_POPULAR_LIST"
            ;;
        go|golang)
            list="$GO_POPULAR_LIST"
            ;;
        maven|java)
            list="$MAVEN_POPULAR_LIST"
            ;;
        *)
            echo "0"
            return
            ;;
    esac

    echo "$list" | wc -l | tr -d ' '
}

#############################################################################
# Typosquatting Detection Functions
#############################################################################

# Calculate Levenshtein distance between two strings
# Usage: levenshtein <string1> <string2>
levenshtein() {
    local s1="$1"
    local s2="$2"

    python3 -c "
def levenshtein(s1, s2):
    if len(s1) < len(s2):
        return levenshtein(s2, s1)
    if len(s2) == 0:
        return len(s1)
    prev_row = range(len(s2) + 1)
    for i, c1 in enumerate(s1):
        curr_row = [i + 1]
        for j, c2 in enumerate(s2):
            insertions = prev_row[j + 1] + 1
            deletions = curr_row[j] + 1
            substitutions = prev_row[j] + (c1 != c2)
            curr_row.append(min(insertions, deletions, substitutions))
        prev_row = curr_row
    return prev_row[-1]
print(levenshtein('$s1', '$s2'))
"
}

# Find similar packages (potential typosquats)
# Usage: find_similar_packages <package> <ecosystem> [threshold]
# Returns: JSON array of similar packages with distances
find_similar_packages() {
    local pkg="$1"
    local ecosystem="${2:-npm}"
    local threshold="${3:-2}"  # Default Levenshtein threshold

    local results="[]"

    # Skip if package is itself popular
    if [[ $(is_popular_package "$pkg" "$ecosystem") == "true" ]]; then
        echo "$results"
        return
    fi

    # Get popular packages for ecosystem
    local popular_list=$(get_popular_packages "$ecosystem")

    while IFS= read -r popular; do
        [[ -z "$popular" ]] && continue

        # Skip if names are identical
        [[ "$pkg" == "$popular" ]] && continue

        # Calculate Levenshtein distance
        local distance=$(levenshtein "$pkg" "$popular")

        # Dynamic threshold based on package name length
        local name_length=${#popular}
        local dynamic_threshold=$threshold
        if [[ $name_length -gt 10 ]]; then
            dynamic_threshold=$((threshold + 1))
        fi

        # Check if within threshold
        if [[ $distance -gt 0 && $distance -le $dynamic_threshold ]]; then
            results=$(echo "$results" | jq --arg pkg "$popular" --argjson dist "$distance" \
                '. + [{"similar_to": $pkg, "distance": $dist}]')
        fi
    done <<< "$popular_list"

    # Sort by distance
    echo "$results" | jq 'sort_by(.distance)'
}

# Comprehensive typosquat check
# Usage: check_typosquat_risk <package> <ecosystem>
# Returns: JSON with risk assessment
check_typosquat_risk() {
    local pkg="$1"
    local ecosystem="${2:-npm}"

    # Skip if package is itself popular
    if [[ $(is_popular_package "$pkg" "$ecosystem") == "true" ]]; then
        echo '{"suspicious": false, "reason": "is_popular_package"}'
        return
    fi

    # Find similar packages
    local similar=$(find_similar_packages "$pkg" "$ecosystem")
    local similar_count=$(echo "$similar" | jq 'length')

    if [[ $similar_count -gt 0 ]]; then
        local min_distance=$(echo "$similar" | jq '[.[].distance] | min')
        local risk_level="low"

        if [[ $min_distance -le 1 ]]; then
            risk_level="high"
        elif [[ $min_distance -le 2 ]]; then
            risk_level="medium"
        fi

        echo "{
            \"suspicious\": true,
            \"risk_level\": \"$risk_level\",
            \"similar_packages\": $similar
        }" | jq '.'
    else
        echo '{"suspicious": false, "reason": "no_similar_packages_found"}'
    fi
}

# Check if package is known malicious
# Usage: is_known_malicious <package> <ecosystem>
is_known_malicious() {
    local pkg="$1"
    local ecosystem="${2:-npm}"
    local list=""

    case "$ecosystem" in
        npm|node)
            list="$NPM_MALICIOUS_LIST"
            ;;
        pypi|python)
            list="$PYPI_MALICIOUS_LIST"
            ;;
        *)
            echo "false"
            return
            ;;
    esac

    if echo "$list" | grep -qx "$pkg"; then
        echo "true"
    else
        echo "false"
    fi
}

#############################################################################
# Export Functions
#############################################################################

export -f is_popular_package
export -f get_popular_packages
export -f get_popular_package_count
export -f levenshtein
export -f find_similar_packages
export -f check_typosquat_risk
export -f is_known_malicious
