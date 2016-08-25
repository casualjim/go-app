#!/bin/bash
set -e -o pipefail

 mkdir -p /drone/{testresults,coverage,dist}
 go test -race -timeout 20m -v $(go list ./... | grep -v vendor) | go-junit-report -dir /drone/testresults

# Run test coverage on each subdirectories and merge the coverage profile.
echo "mode: ${GOCOVMODE-count}" > profile.cov

# Standard go tooling behavior is to ignore dirs with leading underscores
# skip generator for race detection and coverage
for dir in $(go list ./... | grep -v vendor)
do
  go test -covermode=${GOCOVMODE-count} -coverprofile=/drone/src/github.com/casualjim/go-app/profile.out $dir
  if [ -f /drone/src/github.com/casualjim/go-app/profile.out ]
  then
      cat /drone/src/github.com/casualjim/go-app/profile.out | tail -n +2 >> profile.cov
      # rm $pth/profile.out
  fi
done

go tool cover -func profile.cov
gocov convert profile.cov | gocov report
gocov convert profile.cov | gocov-html > /drone/coverage/coverage-${CI_BUILD_NUM-"0"}.html