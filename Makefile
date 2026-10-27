.PHONY: bup, clean, gpush

bup:

clean:

gpush:
	git add ${dir} go.mod go.sum
	git commit -m "${msg}"
	git push
	