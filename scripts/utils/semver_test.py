import timeit
import unittest
from random import shuffle
from typing import NamedTuple

import semver


class Test(NamedTuple):
    input: str
    output: str


_tests = (
    ("bad", ""),
    ("v1-alpha.beta.gamma", ""),
    ("v1-pre", ""),
    ("v1+meta", ""),
    ("v1-pre+meta", ""),
    ("v1.2-pre", ""),
    ("v1.2+meta", ""),
    ("v1.2-pre+meta", ""),
    ("v1.0.0-alpha", "v1.0.0-alpha"),
    ("v1.0.0-alpha.1", "v1.0.0-alpha.1"),
    ("v1.0.0-alpha.beta", "v1.0.0-alpha.beta"),
    ("v1.0.0-beta", "v1.0.0-beta"),
    ("v1.0.0-beta.2", "v1.0.0-beta.2"),
    ("v1.0.0-beta.11", "v1.0.0-beta.11"),
    ("v1.0.0-rc.1", "v1.0.0-rc.1"),
    ("v1", "v1.0.0"),
    ("v1.0", "v1.0.0"),
    ("v1.0.0", "v1.0.0"),
    ("v1.2", "v1.2.0"),
    ("v1.2.0", "v1.2.0"),
    ("v1.2.3-456", "v1.2.3-456"),
    ("v1.2.3-456.789", "v1.2.3-456.789"),
    ("v1.2.3-456-789", "v1.2.3-456-789"),
    ("v1.2.3-456a", "v1.2.3-456a"),
    ("v1.2.3-pre", "v1.2.3-pre"),
    ("v1.2.3-pre+meta", "v1.2.3-pre"),
    ("v1.2.3-pre.1", "v1.2.3-pre.1"),
    ("v1.2.3-zzz", "v1.2.3-zzz"),
    ("v1.2.3", "v1.2.3"),
    ("v1.2.3+meta", "v1.2.3"),
    ("v1.2.3+meta-pre", "v1.2.3"),
    ("v1.2.3+meta-pre.sha.256a", "v1.2.3"),
)
tests = [Test(*t) for t in _tests]


class TestSemver(unittest.TestCase):
    def test_is_valid(self):
        for test in tests:
            self.assertEqual(semver.is_valid(test.input), test.output != "")
            self.assertNotEqual(semver.is_valid(test.input), test.output == "")

    def test_canonical(self):
        for test in tests:
            self.assertEqual(semver.canonical(test.input), test.output)

    def test_major(self):
        for test in tests:
            want = test.output.split(".")[0] if "." in test.output else ""
            self.assertEqual(semver.major(test.input), want)

    def test_major_minor(self):
        for test in tests:
            out = semver.major_minor(test.input)
            want = ""
            if out:
                want = test.input
                if "+" in want:
                    want = want[: want.index("+")]
                if "-" in want:
                    want = want[: want.index("-")]
                match want.count("."):
                    case 0:
                        want += ".0"
                    case 1:
                        pass
                    case 2:
                        want = want[: want.rindex(".")]
                    case _:
                        msg = f"Invalid version: {want}"
                        raise ValueError(msg)
            self.assertEqual(out, want, f"MajorMinor({test.input}) = {out}, want {want}")

    def test_prerelease(self):
        for test in tests:
            pre = semver.prerelease(test.input)
            want = ""
            if test.output != "" and "-" in test.output:
                want = test.output[test.output.index("-") :]
            self.assertEqual(pre, want, f"Prerelease({test.input}) = {pre}, want {want}")

    def test_build(self):
        for test in tests:
            build_suffix = semver.build(test.input)
            want = ""
            if test.output != "" and "+" in test.input:
                want = test.input[test.input.index("+") :]
            self.assertEqual(
                build_suffix, want, f"Build({test.input}) = {build_suffix}, want {want}"
            )

    def test_compare(self):
        for i, ti in enumerate(tests):
            for j, tj in enumerate(tests):
                cmp = semver.compare(ti.input, tj.input)
                want = 0 if ti.output == tj.output else -1 if i < j else 1
                self.assertEqual(
                    cmp, want, f"Compare({ti.input}, {tj.input}) = {cmp}, want {want}"
                )

    def test_sort(self):
        versions = [test.input for test in tests]
        shuffle(versions)
        versions.sort(key=semver.canonical)
        self.assertEqual(versions, sorted(versions, key=semver.canonical), "list is not sorted")


def benchmark_compare():
    v1 = "v1.0.0+metadata-dash"
    v2 = "v1.0.0+metadata-dash1"

    start_time = timeit.default_timer()
    for _ in range(1000000):
        if semver.compare(v1, v2) != 0:
            msg = "bad compare"
            raise ValueError(msg)
    elapsed = timeit.default_timer() - start_time
    print(f"Benchmark completed in: {elapsed} seconds")


def benchmark_compare_prerelease():
    v1 = "v1.2.3-pre"
    v2 = "v1.2.3-pre"

    start_time = timeit.default_timer()
    for _ in range(1000000):
        if semver.compare_prerelease(v1, v2) != 0:
            msg = "bad compare"
            raise ValueError(msg)
    elapsed = timeit.default_timer() - start_time
    print(f"Benchmark completed in: {elapsed} seconds")


def benchmark_next_ident():
    test_string = "v1.0.0-alpha.beta.gamma.delta"
    start_time = timeit.default_timer()
    for _ in range(100000):
        for test in tests:
            semver.next_ident(test.input)
        semver.next_ident(test_string)
    elapsed = timeit.default_timer() - start_time
    print(f"Benchmark completed in: {elapsed} seconds")


if __name__ == "__main__":
    unittest.main()
