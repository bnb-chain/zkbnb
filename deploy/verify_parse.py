import sys

if __name__ == "__main__":
    """
    if (len(sys.argv) < 3):
        print("usage: python3 verify_parse.py src_verifier.sol dest_verifier.sol")
        exit(1)
    """
    src_filename = sys.argv[1]
    dest_filename = sys.argv[2]

    vks = []
    with open(src_filename, "r") as f:
        lines = f.readlines()
        for nu in range(len(lines)):
            if lines[nu].count("function verifyingKey()"):
                for i in range(8):
                    tmp = lines[nu + 1 + i].split("uint256")
                    for j in range(len(tmp) - 1):
                        vks.append("".join([x for x in tmp[j+1] if x.isdigit()]))
                break

    lines = []
    with open(dest_filename, "r") as f:
        lines = f.readlines()
        for nu in range(len(lines)):
            if lines[nu].count("function verifyingKey()"):
                for i in range(14):
                    lines[nu + 1 + i] = "        vk[" + str(i) + "] = " + vks[i] + ";\n"
                    print(lines[nu + 1 + i])

            if lines[nu].count("function ic()"):
                for i in range(8):
                    lines[nu + 2 + i] = "        gammaABC[" + str(i) + "] = " + vks[14 + i] + ";\n"
                    print(lines[nu + 2 + i])

    with open(dest_filename, "w") as f:
        f.writelines(lines)
