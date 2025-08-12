#!/usr/bin/env bash
#Define color variables
#Feature
C_NORMAL='\033[0m';C_BOLD='\033[1m';C_DIM='\033[2m';C_UNDER='\033[4m';
C_ITALIC='\033[3m';C_NOITALIC='\033[23m';C_BLINK='\033[5m';
C_REVERSE='\033[7m';C_CONCEAL='\033[8m';C_NOBOLD='\033[22m';
C_NOUNDER='\033[24m';C_NOBLINK='\033[25m';

#Front color
C_BLACK='\033[30m';C_RED='\033[31m';C_GREEN='\033[32m';C_YELLOW='\033[33m';
C_BLUE='\033[34m';C_MAGENTA='\033[35m';C_CYAN='\033[36m';C_WHITE='\033[37m';

#background color
C_BBLACK='\033[40m';C_BRED='\033[41m';
C_BGREEN='\033[42m';C_BYELLOW='\033[43m';
C_BBLUE='\033[44m';C_BMAGENTA='\033[45m';
C_BCYAN='\033[46m';C_BWHITE='\033[47m';

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color