#!/usr/bin/env bash

# 添加 shell 自动补全
# 使用方法: ./add-completion.sh <命令> <shell>
# 例如: ./add-completion.sh kratos bash
# 参数:
#   cmd: 命令名称
#   shell: 补全文件后缀

cmd=$1
shell=$2

function add()
{
  cat << EOF >> ${HOME}/.bashrc

# ${cmd} shell completion
if [ -f \${HOME}/.${cmd}-completion.bash ]; then
    source \${HOME}/.${cmd}-completion.bash
fi
EOF
}

${cmd} completion ${shell} > ${HOME}/.${cmd}-completion.bash

if ! grep -q "# ${cmd} shell completion" ${HOME}/.bashrc;then
  add
fi
