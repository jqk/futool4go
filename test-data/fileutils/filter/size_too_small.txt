nothing is this file 
'
- filepath.WalkDir() 只在遍历到子目录时调用回调函数
- 如果某个目录下没有子目录,回调函数根本不会被调用
- filepath.Walk() 在遍历每个目录(包括叶子目录)时都会调用回调函数
例如目录结构:'