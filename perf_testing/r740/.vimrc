set nocompatible
set nu
set laststatus=2
set hls
set path+=**
set wildmenu
set showcmd
set binary
set noeol

set t_Co=256
set background=dark

execute pathogen#infect()

syntax on
syntax enable
filetype plugin on
filetype indent on
filetype plugin indent on

set autoindent
set nobackup

colorscheme desert
