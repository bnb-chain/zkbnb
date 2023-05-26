import sys


def find_and_delete_func(func_name, lines):
    first_line = 0
    last_line = 0
    for nu in range(len(lines)):
        if lines[nu].count(func_name):
            first_line = nu
            current_brace = 1
            for j in range(first_line + 1, len(lines)):
                if lines[j].count("{"):
                    current_brace += 1
                if lines[j].count("}"):
                    current_brace -= 1
                if current_brace == 0:
                    last_line = j
                    break
            break
    lines = lines[:first_line] + lines[last_line + 1:]
    print("function name ", func_name, "first line ", first_line, " last line ", last_line)
    return lines, first_line


def getPoint(s):
    return s[s.find("(") + 1:s.find(")")]


if __name__ == "__main__":
    """
    if (len(sys.argv) < 3):
        print("usage: python3 desert_verifier_parse.py src_desert_verifier.sol dest_desert_verifier.sol")
        exit(1)
    """
    src_filename = sys.argv[1]
    dest_filename = sys.argv[2]

    vks = []
    ics = []
    with open(src_filename[i], "r") as f:
        lines = f.readlines()
        for nu in range(len(lines)):
            s = lines[nu]
            if s.count("function verifyingKey()"):
                for i in range(6):
                    tmp = lines[nu + 1 + i].split("uint256")
                    for j in range(len(tmp) - 1):
                        vks.append("".join([x for x in tmp[j + 1] if x.isdigit()]))
            if "vk_x.X = " in s or "mul_input[0] = " in s:
                ic_0 = getPoint(s)
                ic_1 = getPoint(lines[nu + 1])
                ics.append(ic_0)
                ics.append(ic_1)

    lines = []
    with open(dest_filename, "r") as f:
        lines = f.readlines()

        lines, first_line = find_and_delete_func("function verifyingKey()", lines)
        lines, _ = find_and_delete_func("function ic()", lines)
        new_lines = []
        new_lines.append("    function verifyingKey() internal pure returns (uint256[14] memory vk) {\n")
        for j in range(14):
            new_lines.append("            vk[" + str(j) + "] = " + vks[j] + ";\n")
        new_lines.append("            return vk;\n")
        new_lines.append("    }\n\n")

        new_lines.append("    function ic() internal pure returns (uint256[] memory gammaABC) {\n")
        new_lines.append("            gammaABC = new uint256[](4);\n")
        for j in range(4):
            new_lines.append("            gammaABC[" + str(j) + "] = " + ics[j] + ";\n")
        new_lines.append("            return gammaABC;\n")
        new_lines.append("    }\n\n")
        print(new_lines)
        update_lines = lines[:first_line + 1] + new_lines + lines[first_line + 1:]
        """
        for nu in range(len(lines)):
            if lines[nu].count("function verifyingKey()"):
                for i in range(14):
                    lines[nu + 1 + i] = "        vk[" + str(i) + "] = " + vks[i] + ";\n"
                    print(lines[nu + 1 + i])

            if lines[nu].count("function ic()"):
                for i in range(8):
                    lines[nu + 2 + i] = "        gammaABC[" + str(i) + "] = " + vks[14 + i] + ";\n"
                    print(lines[nu + 2 + i])
        """
    with open(dest_filename, "w") as f:
        f.writelines(update_lines)
