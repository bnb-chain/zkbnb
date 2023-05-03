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
        print("usage: python3 verify_parse.py src_verifier1.sol,src_verifier10.sol, 1,10 dest_verifier.sol")
        exit(1)
    """
    src_filename_str = sys.argv[1]
    src_block_size_str = sys.argv[2]
    dest_filename = sys.argv[3]

    src_filenames = sys.argv[1].split(",")
    src_block_sizes = sys.argv[2].split(",")
    if len(src_filenames) != len(src_block_sizes):
        print("source filenames not match block sizes")
        exit(1)

    all_vks = []
    all_ics = []
    for i in range(len(src_filenames)):
        vks = []
        ics = []
        with open(src_filenames[i], "r") as f:
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
        all_vks.append(vks)
        all_ics.append(ics)

    lines = []
    with open(dest_filename, "r") as f:
        lines = f.readlines()

        lines, first_line = find_and_delete_func("function verifyingKey(uint16 block_size)", lines)
        lines, _ = find_and_delete_func("function ic(uint16 block_size)", lines)
        new_lines = []
        new_lines.append("    function verifyingKey(uint16 block_size) internal pure returns (uint256[14] memory vk) {\n")
        for i in range(len(src_block_sizes)):
            if i == 0:
                new_lines.append("        if (block_size == " + src_block_sizes[i] + ") {\n")
            else:
                new_lines.append("        } else if (block_size == " + src_block_sizes[i] + ") {\n")
            for j in range(14):
                new_lines.append("            vk[" + str(j) + "] = " + all_vks[i][j] + ";\n")
            new_lines.append("            return vk;\n")
        new_lines.append("        } else {\n")
        new_lines.append("            revert(\"u\");\n")
        new_lines.append("        }\n")
        new_lines.append("    }\n\n")

        new_lines.append("    function ic(uint16 block_size) internal pure returns (uint256[] memory gammaABC) {\n")
        for i in range(len(src_block_sizes)):

            if i == 0:
                new_lines.append("        if (block_size == " + src_block_sizes[i] + ") {\n")
            else:
                new_lines.append("        } else if (block_size == " + src_block_sizes[i] + ") {\n")
            new_lines.append("            gammaABC = new uint256[](4);\n")
            for j in range(4):
                new_lines.append("            gammaABC[" + str(j) + "] = " + all_ics[i][j] + ";\n")
            new_lines.append("            return gammaABC;\n")
        new_lines.append("        } else {\n")
        new_lines.append("            revert(\"u\");\n")
        new_lines.append("        }\n")
        new_lines.append("    }\n\n")
        print(new_lines)
        update_lines = lines[:first_line + 1] + new_lines + lines[first_line + 1:]
        """
        for nu in range(len(lines)):
            if lines[nu].count("function verifyingKey(uint16 block_size)"):
                for i in range(14):
                    lines[nu + 1 + i] = "        vk[" + str(i) + "] = " + vks[i] + ";\n"
                    print(lines[nu + 1 + i])

            if lines[nu].count("function ic(uint16 block_size)"):
                for i in range(8):
                    lines[nu + 2 + i] = "        gammaABC[" + str(i) + "] = " + vks[14 + i] + ";\n"
                    print(lines[nu + 2 + i])
        """
    with open(dest_filename, "w") as f:
        f.writelines(update_lines)
