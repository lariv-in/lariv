# Caveats when working on this codebase

- In nearly all cases, we should take the address of components before inserting into something that requires PageInterface. Not doing so means that the value will not implement MutableParentInterface and its children won't be patchable
