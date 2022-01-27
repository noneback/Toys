for i in {1..50}
do
	echo "Test: No.${i}: go test -run ${1} ${2}"

    result=$(go test -run ${1} ${2})
	num=$(echo "${result}"| grep "FAIL" | wc -l)
  ZERO=0
#  echo ${num}
	if test ${ZERO} -eq ${num}
	then
    echo "PASS"
        else
    echo "FAIL"
    echo "${result}"
    exit
	fi
done
echo "finished"