.PHONY: bup, clean, gpush

bup:

clean:

gpush:
	git add ${dir} backend/go.mod backend/go.sum
	git commit -m "${msg}"
	git push
