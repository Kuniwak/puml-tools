# puml-tools (PlantUML Tools)

A suite of tools for viewing, editing and searching PlantUML diagrams.

## Installation
### Homebrew

```console
$ brew install kuniwak/tap/puml-tools
```

### Binary download
Download the latest binaries from [Releases](https://github.com/Kuniwak/puml-tools/releases) and place them somewhere on your `PATH`.


## Breaking changes in 3.0.0

- The sequence-diagram-only tools were renamed to share a `pumlseq` prefix. Replace the old names in your scripts.
  - `pumlgroupcut` → `pumlseqgroupcut`
  - `pumlgroupmask` → `pumlseqgroupmask`
- `pumlhelp` now lists tools in lexicographic order by tool name instead of definition order. Scripts that depend on the line order of `pumlhelp -short` are affected.


## Usage examples

```console
$ # Extract PlantUML source from an image rendered by PlantUML.
$ pumlparse <./input.puml.png
@startuml
...
@enduml

$ # If it is already PlantUML text, it is shown as-is.
$ pumlparse ./input.puml.png
@startuml
...
@enduml

$ # Cut out only the specified group block from a sequence diagram (keeping only the participants it references).
$ pumlseqgroupcut -exact Auth ./seq.puml
@startuml
participant B
participant C
group Auth
  B -> C: verify
  C -> B: ok
end
@enduml

$ # Remove the specified group block from a sequence diagram (also removing any participant used only inside that group).
$ pumlseqgroupmask -exact Auth ./seq.puml
@startuml
participant A
participant B
A -> B: login
B -> A: done
@enduml
```

```console
$ # Split a sequence diagram whose messages are all wrapped in groups into one *.puml per group.
$ pumlseqgroupsplit -o ./out ./seq.puml
./out/seq_1_Login.puml
./out/seq_2_Order.puml
```

```console
$ # Enumerate a sequence diagram's event sequences (traces) and write each one to its own TSV.
$ pumlseqtrace -o ./out ./seq.puml
./out/seq_trace_1.tsv
./out/seq_trace_2.tsv

$ cat ./out/seq_trace_1.tsv
login
ok
```

```console
$ # Extract TODOs from PlantUML comments/notes into a TSV listing.
$ pumltodo ./diagram.puml
./diagram.puml	3	TODO: handle the error case
./diagram.puml	8	TODO implement the reply

$ # PNGs are also supported: PlantUML embedded in the metadata is parsed and extracted.
$ pumltodo ./diagram.puml.png
./diagram.puml.png	3	TODO: handle the error case

$ # --regexp lets you match things other than TODO (e.g. FIXME).
$ pumltodo --regexp '(TODO|FIXME)' ./diagram.puml
./diagram.puml	3	TODO: handle the error case
./diagram.puml	12	FIXME: fix this
```

```console
$ # Detect messages in a sequence diagram whose causal chain is broken (the sender has no preceding trigger).
$ pumlseqbreak ./diagram.puml
./diagram.puml	3	sender c has no preceding trigger

$ # Returns exit code 1 when one or more breaks are found, so it can fail a CI build.
$ pumlseqbreak ./diagram.puml; echo $?
1
```

```console
$ # Check whether two sequence diagrams represent the same message sequence.
$ # (Each message is treated as a (sender, receiver) pair; the message body is ignored.)
$ cat ./super.puml          # the complete version
@startuml
A -> B: login
B -> C: verify
C -> B: ok
B -> A: done
@enduml
$ cat ./sub.puml            # the (possibly incomplete) diagram supplied by the customer
@startuml
A -> B: login
B -> X: extra
B -> A: done
@enduml

$ # The default "exact" mode lists the surplus on both sides and returns exit code 1 if there is any.
$ # The first column shows which diagram the message was only found in.
$ pumlseqsubset ./super.puml ./sub.puml; echo $?
sub	B	X
super	B	C
super	C	B
1

$ # If the message sequences are identical, nothing is printed and it exits 0 (so CI can fail on a mismatch).
$ pumlseqsubset ./super.puml ./super.puml; echo $?
0

$ # allow-subseq allows a surplus on the super side and lists only what sub is missing.
$ # (Useful for checking whether an incomplete diagram is contained in the complete one.)
$ # The residual is always on the sub side, so no "which side" column is printed.
$ pumlseqsubset -mode allow-subseq ./super.puml ./sub.puml; echo $?
B	X
1

$ # par branches are concurrent, so they are treated as the same sequence even if listed in a different order than written.
$ cat ./par_super.puml
@startuml super
participant A
participant B
participant C
par
    A -> B: notify B
else
    A -> C: notify C
end
@enduml
$ cat ./par_sub.puml
@startuml sub
participant A
participant B
participant C
A -> C: notify C
A -> B: notify B
@enduml
$ pumlseqsubset ./par_super.puml ./par_sub.puml; echo $?
0
```

```console
$ # Fold several participants of a sequence diagram into one (a summary — the reverse of refinement).
$ pumlseqfold -ids B,C -new-id BE -new-desc "Backend" ./seq.puml
@startuml
participant A
participant "Backend" as BE
A -> BE: req
BE -> BE: fwd
BE -> BE: resp
BE -> A: resp
@enduml

$ # Extract the "shared (synchronizing) events" of the participant set to be folded — i.e. messages whose both endpoints are in that set.
$ pumlseqmsg -common B,C ./seq.puml
B	C	fwd
C	B	resp

$ # Hide those shared events first, then fold: the internal exchange disappears and the summary is complete.
$ pumlseqhide --msg-file <(pumlseqmsg -common B,C ./seq.puml) ./seq.puml \
  | pumlseqfold -ids B,C -new-id BE -new-desc "Backend"
@startuml
participant A
participant "Backend" as BE
A -> BE: req
BE -> A: resp
@enduml

$ # Being TSV, you can insert TSV-processing tools in between (hide reads the set from stdin via --msg-file -).
$ pumlseqmsg -common B,C ./seq.puml | pumlseqhide --msg-file - ./seq.puml | pumlseqfold -ids B,C -new-id BE -new-desc "Backend"

$ # A branch that consisted only of internal messages becomes empty after hide, but empty blocks are removed
$ # automatically, and a par/alt left with a single branch is inlined (its indentation is flattened to the parent level).
$ cat ./inner.puml
@startuml
participant A
participant B
participant C
A -> B: req
par
    A -> B: notify
else
    B -> C: sync
    C -> B: ack
end
B -> A: resp
@enduml
$ pumlseqmsg -common B,C ./inner.puml | pumlseqhide --msg-file - ./inner.puml
@startuml
participant A
participant B
participant C
A -> B: req
A -> B: notify
B -> A: resp
@enduml
```

```console
$ # Restrict to a specific participant's viewpoint (projection — trace restriction tr ↾ A; the dual of pumlseqhide).
$ # Messages that don't touch C are dropped, along with declarations (participant A) that are no longer referenced.
$ pumlseqproj -ids C ./seq.puml
@startuml
participant B
participant C
B -> C: fwd
C -> B: resp
@enduml

$ # Internal events among the retained participant set are not dropped by projection (exactly opposite of hiding).
$ pumlseqproj -ids B,C ./seq.puml | pumlseqmsg -common B,C -
B	C	fwd
C	B	resp
```

```console
$ # Check whether "detail" is a refinement of "abstract" (CSP trace refinement).
$ # -map supplies the mapping that folds detail's participants into abstract's participants.
$ cat ./map.tsv
B	BE
C	BE

$ # If the folded detail's traces fit within abstract's traces, refinement holds (exit code 0, no output).
$ pumlseqrefinement -map ./map.tsv ./abstract.puml ./detail.puml; echo $?
0

$ # Internal messages inside the box remain visible, so if there is internal traffic not present in abstract,
$ # refinement does not hold (exit code 1).
$ pumlseqrefinement -map ./map.tsv ./abstract.puml ./detail-internal.puml; echo $?
Refinement does not hold.
After the following trace, detail can perform an event that abstract does not allow:
  trace: A -> BE: req
  forbidden event: BE -> BE: secret
1

$ # If you want to ignore the internal exchange on the abstract side, hide it first with pumlseqhide, then pass it in (- reads from stdin).
$ pumlseqhide --msg-file <(pumlseqmsg -common B,C ./detail-internal.puml) ./detail-internal.puml \
  | pumlseqrefinement -map ./map.tsv ./abstract.puml -; echo $?
0
```

```console
$ # "ref over" references another diagram. Write the label as a path relative to the referencing diagram.
$ cat ./seq.puml
@startuml
participant A
participant B
A -> B: login
ref over A, B : ./auth.puml
@enduml
$ cat ./auth.puml
@startuml
participant A
participant B
A -> B: verify
B -> A: ok
@enduml

$ # pumlseqexpand replaces the ref with the referenced body (recursively; the ref line's indentation carries over to the body).
$ pumlseqexpand ./seq.puml
@startuml
participant A
participant B
A -> B: login
A -> B: verify
B -> A: ok
@enduml

$ # Inspection tools do not silently ignore an unexpanded ref — they fail — so expand first, then pass the result in.
$ pumlseqsubset ./seq.puml ./sub.puml; echo $?
./seq.puml:5: unexpanded ref found: ref over A, B; run it through pumlseqexpand first
2
$ pumlseqexpand ./seq.puml | pumlseqrefinement -map ./map.tsv ./abstract.puml -; echo $?
0

$ # All of the following exit with code 2: the label is not a relative path; the referenced file doesn't exist;
$ # the reference is circular; the referenced diagram uses a lifeline not listed in "over" (a scope violation);
$ # the referenced file doesn't contain exactly one diagram.
$ # find's -exec doesn't propagate exit codes, so use xargs instead (xargs's own exit code is nonzero when
$ # a child fails: BSD xargs uses 1, GNU xargs uses 123).
$ find . -name '*.puml' -print0 | xargs -0 -n1 pumlseqexpand >/dev/null; echo $?
auth.puml uses lifeline "C", which is not covered by ref over A, B; add it to the over clause or fix the ref target
1
```

```console
$ # Build a single C4-PlantUML diagram from the union of several sequence diagrams (participant → element, message → Rel).
$ # The diagram name is given with -name; if omitted, it is left unnamed (just @startuml).
$ pumlseq2c4 -name MyContext ./seq-01.puml ./seq-02.puml
@startuml MyContext
!include <C4/C4_Context>

System(a, "A")
System(b, "B")
System(c, "C")

Rel(a, b, "fwd")
Rel(b, c, "sync")
@enduml

$ # Check whether a sequence diagram and a C4 diagram are consistent (-method eq is an exact match; 0 if consistent).
$ pumlc42seqchk -method eq -seq ./seq.puml -c4 ./c4.puml; echo $?
OK: consistent
0

$ # -method subset checks whether the C4 diagram is a subset of the sequence diagram, returning 1 if there is any excess or shortfall.
$ pumlc42seqchk -method subset -seq ./seq.puml -c4 ./c4.puml; echo $?
elements only in C4 (not in SEQ): 1
  - foo ("Foo")
1

$ # -method superset checks whether the C4 diagram is a superset of the sequence diagram. This lets you check that
$ # each of several sequence diagrams is covered by a single C4 diagram built from them.
$ for seq in ./seq-*.puml; do pumlc42seqchk -method superset -seq "$seq" -c4 ./c4.puml; done; echo $?
OK: seq ⊆ c4
OK: seq ⊆ c4
0

$ # Sequence diagrams can be given as multiple positional arguments; the SEQ side becomes their union, so you can use
$ # eq to check whether the C4 diagram exactly matches the diagram made by composing several sequence diagrams (no excess or shortfall).
$ pumlc42seqchk -method eq -c4 ./c4.puml ./seq-01.puml ./seq-02.puml; echo $?
OK: consistent
0

$ # When there are many, pass them with find's -exec ... +.
$ find . -name 'seq-*.puml' -exec pumlc42seqchk -method eq -c4 ./c4.puml {} +; echo $?
OK: consistent
0

$ # If a participant with the same id is declared with a different display name or kind across diagrams, they cannot be composed.
$ pumlc42seqchk -method eq -c4 ./c4.puml ./seq-01.puml ./seq-02.puml; echo $?
participant "a" conflicts between ./seq-01.puml and ./seq-02.puml: Person "a" vs System "a"
2
```

```console
$ # Checking across levels of detail: if C4 is a Container diagram (C4 Lv.2, the system split into containers)
$ # and the sequence diagram treats that system as a single unit (C4 Lv.1), fold the containers back into the
$ # system with pumlc4fold before running pumlc42seqchk.
$ # pumlc4bdyls lists the elements to fold as a TSV with a header, keyed by boundary (pumlc4fold reads the id column).
$ pumlc4bdyls -boundary shop ./c4-lv2.puml
id	kind	external	boundary	label
web	Container	false	shop	Web Storefront
order_svc	Container	false	shop	Order Service

$ # Pass the TSV via process substitution to fold (you can also thin it with qhs or tsv-utils before passing it in).
$ pumlc4fold -ids-tsv <(pumlc4bdyls -boundary shop ./c4-lv2.puml) -new-id shop -new-desc "Shop" ./c4-lv2.puml \
  | pumlc42seqchk -method eq -seq ./seq-lv1.puml -c4 -; echo $?
OK: consistent
0
```


## Usage

### pumlc42seqchk

```console
$ pumlc42seqchk -h
Usage: pumlc42seqchk (-seq <path|->|<path|->...) -c4 <path|-> [-method eq|subset|superset]

Checks whether a sequence diagram and a C4 diagram are consistent with each
other. Participants and C4 elements are compared by id, and messages and
Rels are compared as directed (sender, receiver) pairs (labels are ignored,
self-relations are dropped, and boundaries don't count as elements).
-method eq (the default) requires the node sets and directed edge sets to
match exactly; -method subset requires C4 to be a subset of the sequence
diagram (C4 subset-of SEQ); -method superset requires C4 to be a superset
of the sequence diagram (SEQ subset-of C4).
Multiple sequence diagrams can be given as positional arguments. In that
case SEQ is their union (participants deduplicated by id, messages
deduplicated as directed pairs), so you can check whether several sequence
diagrams combined (eq) match a single C4 diagram. It is an error if the
same participant id is declared with a different display name or kind in
different diagrams, since they can't then be merged.
A ref over block does not contribute to C4 (it's a call into another
diagram, so it produces neither elements nor Rels); when one is found, a
warning is printed to stderr. The check still proceeds, but it may report
"inconsistent" purely because the referenced relationships are missing, so
either pass the referenced diagram as an additional positional argument, or
run the input through pumlseqexpand.
Returns exit code 0 if consistent, 1 if inconsistent, and 2 on error. When
inconsistent, the differences are listed on stdout. At most one input may
be - to read from stdin.
For a PNG input, the embedded PlantUML metadata is parsed.

Options:
  -c4 string
    	path to the C4 diagram (- for stdin)
  -method string
    	check method eq|subset|superset (default eq) (default "eq")
  -seq string
    	path to the sequence diagram (- for stdin); mutually exclusive with positional arguments
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlc42seqchk -method subset -seq seq.puml -c4 c4.puml; echo $?
  elements only in C4 (not in SEQ): 1
    - foo ("Foo")
  1

  $ pumlc42seqchk -method eq -c4 c4.puml seq-01.puml seq-02.puml; echo $?
  OK: consistent
  0
```

### pumlc4bdyls

```console
$ pumlc4bdyls -h
Usage: pumlc4bdyls -boundary <id> (<file>|< <file>)

Lists, as a headered TSV, the elements contained (including nested) in the
C4 diagram boundary <id>. The columns are id / kind / external / boundary /
label, where boundary is each element's immediately enclosing boundary id.
Sub-boundary ids themselves are not listed, but the elements inside them
are. The output can be piped directly into pumlc4fold -ids-tsv (filter it
first with your favorite TSV tool if needed).
It is an error, and the command exits with a failure, if -boundary names a
boundary that doesn't exist in the diagram.
For a PNG input, the embedded PlantUML metadata is parsed.

Options:
  -boundary string
    	id of the boundary whose contained elements to list
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlc4bdyls -boundary shop ./c4.puml
  $ pumlc4bdyls -boundary shop <./c4.puml
```

### pumlc4fold

```console
$ pumlc4fold -h
Usage: pumlc4fold (-ids <a,b,...>|-ids-tsv <file>) -new-id <id> [options] (<file>|< <file>)

Folds several elements of a C4 diagram into one. The folded element is
always emitted as a System, so folding the containers of a Container
diagram (C4 Lv.2) back into their system yields a System Context diagram
(C4 Lv.1). Folding an arbitrary subset is allowed too; the output then
keeps whatever kinds were left untouched, and the !include is chosen to
cover the most detailed kind still present.
Every Rel to or from an element listed in -ids is rewritten to point at the
-new-id element instead. Rels between the folded elements (now self-Rels)
are dropped, and Rels that end up with the same (source, destination, label)
are merged into one. The output is a flat C4 diagram with no boundaries.
The elements to fold can be given directly with -ids (comma-separated), or
read from the id column of a headered TSV with -ids-tsv (intended for
piping in the output of pumlc4bdyls via process substitution; filter it
first with your favorite TSV tool if needed). -ids and -ids-tsv are
mutually exclusive, and exactly one is required. It is an error to name an
element that doesn't exist in the diagram, and an error if -new-id collides
with an existing element that isn't being folded. Pass -name to attach a
name to the output's @startuml (omit it to keep the input's name). For a
PNG input, the embedded PlantUML metadata is parsed.

Options:
  -ids string
    	ids of the elements to fold (comma-separated, one or more)
  -ids-tsv string
    	path to a headered TSV whose id column lists the elements to fold (mutually exclusive with -ids)
  -name string
    	name to attach to the output's @startuml (defaults to the input's name)
  -new-desc string
    	display name of the folded element (optional; defaults to new-id)
  -new-id string
    	id of the folded element
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlc4fold -ids web,order_svc -new-id shop -new-desc "Shop" ./c4.puml
  $ pumlc4fold -ids-tsv <(pumlc4bdyls -boundary shop ./c4.puml) -new-id shop ./c4.puml
```

### pumlhelp

```console
$ pumlhelp -h
Usage: pumlhelp [options] [tool ...]

Prints the help of every PlantUML tool, gathered together as Markdown.
With no arguments it shows every tool; naming one or more tools shows only those.
Tools are always listed in dictionary order by name, not in the order given.

Options:
  -short
    	print only a one-line summary of each tool
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlhelp
  $ pumlhelp -short
  $ pumlhelp pumlparse
```

### pumlparse

```console
$ pumlparse -h
Usage: pumlparse [options] (<file>|< <file>)

Extracts PlantUML source embedded in a PNG's metadata, parses it into an AST, and
renders it back to text (a round trip). If the input file is already plain text,
it is parsed and rendered back the same way.
With --raw, the extracted PlantUML is printed as-is, without parsing into an AST.

Options:
  -raw
    	print the PlantUML embedded in the metadata as-is, without parsing it into an AST
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlparse <./input.puml.png
  @startuml
  ...
  @enduml

  $ pumlparse ./input.puml.png
  @startuml
  ...
  @enduml
```

### pumlseq2c4

```console
$ pumlseq2c4 -h
Usage: pumlseq2c4 [options] (<file>|-)...

Reads one or more PlantUML sequence diagrams and builds a single
C4-PlantUML diagram from their union. Each participant becomes a C4
element (actor becomes Person, everything else becomes System), and each
message becomes a Rel (with the message body as its label). With
multiple inputs, elements are deduplicated by id and Rels by (sender,
receiver, label) across inputs, and self-messages are dropped. No
diagram name is set (pass -name to set one). Aggregating or renaming
participants is out of scope; inputs are assumed to already correspond
1:1 with C4 (use pumlseqfold beforehand if you need to merge them). A
ref over block does not contribute to C4 (it's a call into another
diagram, so it produces neither elements nor Rels); when one is found, a
warning is printed to stderr. Either pass the referenced diagram as an
additional positional argument, or run the input through pumlseqexpand.
Passing -boundary box turns each sequence diagram's box into a C4
System_Boundary, placing the participants declared inside that box into
the boundary. Participants that don't belong to a box stay outside any
boundary. With multiple inputs, it is an error if the same participant
belongs to a different box in different diagrams.
Omit the arguments to read a single input from stdin. Either argument may
be - for stdin. For a PNG input, the embedded PlantUML metadata is
parsed.

Options:
  -boundary string
    	how to build boundaries: none emits no boundaries (default), box turns each sequence diagram's box into a System_Boundary (default "none")
  -name string
    	name of the output C4 diagram (attached to @startuml); no name if omitted
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseq2c4 ./seq-01.puml ./seq-02.puml
  @startuml
  !include <C4/C4_Context>

  System(a, "A")
  System(b, "B")

  Rel(a, b, "fwd")
  @enduml
```

### pumlseqbreak

```console
$ pumlseqbreak -h
Usage: pumlseqbreak [options] (<file>|< <file>)

Scans a sequence diagram's message arrows and reports each message whose
sender has no preceding trigger (a break in the causal chain), listing
them as a 3-column TSV of file name, line number, and reason.
Each message is checked only against the immediately preceding message
(a chain a->b/b->c or a broadcast a->b/a->c is fine).
A "..." line resets causality and always permits the next message.
A ref over is a call into another diagram, so it is treated as a single
opaque event involving its target lifelines (after the ref, a send from
any of its targets counts as a trigger; a ref whose targets have no
trigger at all is itself reported as a break).
Blocks such as alt/opt/loop/group/par are checked against the context in
effect at entry, once per branch.
Exit status is 1 if one or more breaks are found. For PNG input, the
embedded PlantUML metadata is read.

Options:
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqbreak ./diagram.puml
  ./diagram.puml	3	sender c has no preceding trigger
```

### pumlseqexpand

```console
$ pumlseqexpand -h
Usage: pumlseqexpand [options] (<file>|< <file>)

Replaces each ref over in a sequence diagram with the body of the diagram it
references (recursively). A ref's label must be a path relative to the
referencing file. Everything below is an error (exit code 2): a label that
isn't a relative path, a target that can't be read, a reference cycle, a
scope violation (the target uses a lifeline missing from the ref's over
clause), a participant declaration that conflicts across diagrams, and a
target that isn't a single diagram enclosed by @startuml and @enduml. Run it
over every diagram in CI to catch mistakes in how refs are written (find's
-exec doesn't propagate exit codes, so pipe through xargs as in the example
below).

A ref target must be a single diagram enclosed by @startuml and @enduml. Any
diagram-wide directive the target carries for rendering on its own
(title/caption/header/footer/legend/autonumber/skinparam/skin/scale/hide/
show/newpage) is dropped instead of inlined, and each drop is reported to
stderr as a warning (the parent diagram's own directive still applies, so its
appearance is unchanged). A directive written across multiple lines
(title/header/footer/legend) is dropped along with its whole body, up to
"end <keyword>". Comments, blank lines, and dividers (== x ==) are positional
and are always inlined as-is; the contents of a block comment are left alone
since they aren't a diagram directive.

A target's participant declarations are not left at the expansion site;
only the ones not already declared are appended to the end of the parent
diagram's declaration block. A declaration that was inside a box in the
target goes into the parent's box of the same name if one exists, otherwise
the whole box is appended to the end of the declaration block. Declarations
must agree across diagrams: declaring the same id with a different kind or
label is an error, and so is disagreeing about which box it belongs to (or
whether it's outside any box at all). A box models which lifelines belong
together, so diagrams joined by ref must place the same participant in the
same box. Errors and warnings are formatted as
"<referencing file>:<ref's line>: ...", so it's clear which ref is at fault
even when processing many diagrams at once.

When given a file argument, relative paths are resolved from that file's
directory (combining this with -base is an error). When reading from stdin,
use -base to set that starting directory (it defaults to the current
directory). A ref's label can't contain whitespace, so files with whitespace
in their names can't be referenced. The output's @startuml keeps the input's
name; pass -name to override it. Note: inlined bodies and hoisted
declarations carry over the target's own line terminators as-is, so the
output will mix CRLF and LF if the parent and target use different ones
(this doesn't affect how PlantUML interprets it).

Options:
  -base string
    	directory ref's relative paths are resolved from when reading stdin (default: current directory)
  -name string
    	diagram name to give the output's @startuml (default: keeps the input's name)
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqexpand ./seq.puml
  @startuml
  participant A
  participant B
  A -> B: req
  A -> B: inner
  @enduml

  $ pumlseqexpand ./seq.puml | pumlseqrefinement -map ./map.tsv ./abstract.puml -

  $ find . -name '*.puml' -print0 | xargs -0 -n1 pumlseqexpand >/dev/null
```

### pumlseqfold

```console
$ pumlseqfold -h
Usage: pumlseqfold -ids <a,b,...> -new-id <id> [options] (<file>|< <file>)

Folds several participants of a sequence diagram into one (summarization). Every message and
reference (activate/deactivate, note over, ref over) addressed to or from a participant in
-ids is rewritten to belong to the -new-id participant. If folding collapses a ref over's
targets into one, duplicates are removed.
Messages between folded participants remain as self-transitions (remove them with pumlseqhide).
It is an error if -ids names a participant that does not exist in the diagram, and likewise if
-new-id collides with an existing participant outside the fold.
Control blocks left empty by folding or a prior hide (empty alt/par branches, or empty
opt/loop/group/break, etc.) are pruned automatically, and a par/alt reduced to a single branch
has its contents hoisted to the parent level (recursively, from the innermost block out, and
idempotently).
Pass -name to set the name on the output's @startuml (omit it to keep the input's name).
For PNG input, the embedded PlantUML metadata is decoded.
Note: even when the input uses CRLF line endings, newly inserted participant declaration lines
end with LF (the mixed line endings do not affect how PlantUML interprets the diagram).

Options:
  -ids string
    	participant ids to fold (comma separated, at least one)
  -name string
    	name to set on the output's @startuml (omit to keep the input's name)
  -new-desc string
    	description for the resulting folded participant (optional; omit for no description)
  -new-id string
    	id of the resulting folded participant
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqfold -ids B,C -new-id BE -new-desc "Backend" ./seq.puml
  $ pumlseqfold -ids B,C -new-id BE -name Folded <./seq.puml
```

### pumlseqgroupcut

```console
$ pumlseqgroupcut -h
Usage: pumlseqgroupcut -exact <label> [options] (<file>|< <file>)

Extracts the group block with the given label from a sequence diagram.
Only the declarations of participants referenced inside the extracted group
are kept; every other line (messages outside the group, unreferenced
participant declarations) is dropped. The group ... end frame itself is
kept as-is. If multiple groups match, all of them are processed.
The output @startuml does not carry over the input's diagram name; pass
-name to give it one.

Options:
  -exact string
    	label of the group to extract (exact match against the text following the group keyword)
  -name string
    	diagram name to give the output @startuml (default: no name)
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqgroupcut -exact Auth ./seq.puml
  @startuml
  participant B
  participant C
  group Auth
    B -> C: verify
    C -> B: ok
  end
  @enduml

  $ pumlseqgroupcut -exact Auth -name AuthExcerpt <./seq.puml
  @startuml AuthExcerpt
  ...
  @enduml
```

### pumlseqgroupmask

```console
$ pumlseqgroupmask -h
Usage: pumlseqgroupmask -exact <label> [options] (<file>|< <file>)

Removes the group block with the given label from a sequence diagram.
Declarations of participants that were referenced only inside the removed
group (participants no longer referenced anywhere in the remaining diagram)
are removed as well. Every other line is left as-is.
If multiple groups match, all of them are processed.
The output @startuml does not carry over the input's diagram name; pass
-name to give it one.

Options:
  -exact string
    	label of the group to remove (exact match against the text following the group keyword)
  -name string
    	diagram name to give the output @startuml (default: no name)
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqgroupmask -exact Auth ./seq.puml
  @startuml
  participant A
  participant B
  A -> B: login
  B -> A: done
  @enduml

  $ pumlseqgroupmask -exact Auth -name Overview <./seq.puml
  @startuml Overview
  ...
  @enduml
```

### pumlseqgroupsplit

```console
$ pumlseqgroupsplit -h
Usage: pumlseqgroupsplit -o <out_dir> [options] (<file>|< <file>)

Splits a sequence diagram whose every message is enclosed in a group into
one *.puml file per group. The unit of the split is the outermost group;
a nested group stays inside its enclosing group's file. Each file is cut
out the same way as pumlseqgroupcut (only the declarations of participants
referenced inside that group and the group ... end frame are kept, every
other line is dropped). The output @startuml is left unnamed.
File names follow <out_dir>/<basename>_<idx>_<group title>.puml, where idx
starts at 1 and is zero-padded to the number of digits in the group count.
Spaces in the title and characters not allowed in file names (/ \ : * ? "
< > |) are replaced with "_". A group with an empty title produces
<basename>_<idx>.puml.
basename defaults to the input file name with a trailing .png or .puml
stripped. When reading from stdin, -basename is required, and a -basename
containing a path separator (/ \) is an error.
<out_dir> is created if it does not exist, and files with the same name
are overwritten. Every file is written to a temporary file first and only
then renamed into place, so a failure while generating or writing content
leaves <out_dir> untouched. Once the renames begin they are not rolled
back: if a rename itself fails, the files already renamed stay in place.
Written paths are printed to stdout, one per line.
It is an error if there are no groups at all, or if even one message is
not enclosed in any group (such a message would silently fall out of
every file). ref over outside a group is not checked; if its target is
also enclosed in a group, run pumlseqexpand first so that the expanded
messages end up inside the group before splitting.
For PNG input, the embedded PlantUML metadata is read.

Options:
  -basename string
    	base name for output files (default: derived from the input file name; required when reading from stdin)
  -o string
    	directory to write the split *.puml files into (created if it does not exist)
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqgroupsplit -o ./out ./seq.puml
  ./out/seq_1_Login.puml
  ./out/seq_2_Order.puml

  $ pumlseqgroupsplit -o ./out -basename flow <./seq.puml
  ./out/flow_1_Login.puml
  ./out/flow_2_Order.puml
```

### pumlseqhide

```console
$ pumlseqhide -h
Usage: pumlseqhide -msg-file <path|-> [options] (<file>|< <file>)

Removes, line by line, every message in a sequence diagram that matches the message set given
by -msg-file (this hides the internal interaction as tau, i.e. CSP hiding P \ A). The message
set is a TSV of (sender<TAB>receiver<TAB>label); the output of pumlseqmsg can be piped in
directly.
Matching is on the (sender, receiver, label) triple; arrow style is ignored. Any note attached
to a hidden message (one immediately following it, attached to that participant) is removed
along with it. The note's own text is not part of the match.
ref over is never hidden and is left untouched (to remove an internal ref, first run
pumlseqexpand to turn it into real messages).
Control blocks left empty by the removal (empty alt/par branches, or empty
opt/loop/group/break, etc.) are pruned automatically, and a par/alt reduced to a single branch
has its contents hoisted to the parent level (recursively, from the innermost block out, and
idempotently).
Pass - to -msg-file to read the message set from stdin (in that case, pass the diagram as a
file argument instead).
Pass -name to set the name on the output's @startuml (omit it to keep the input's name).
For PNG input, the embedded PlantUML metadata is decoded.

Options:
  -msg-file string
    	TSV of messages to hide (sender<TAB>receiver<TAB>label); - reads from stdin
  -name string
    	name to set on the output's @startuml (omit to keep the input's name)
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqhide --msg-file <(pumlseqmsg --common B,C ./seq.puml) ./seq.puml
  $ pumlseqmsg --common B,C ./seq.puml | pumlseqhide --msg-file - ./seq.puml
  $ pumlseqhide --msg-file - -name Hidden ./seq.puml <./msgs.tsv
```

### pumlseqmsg

```console
$ pumlseqmsg -h
Usage: pumlseqmsg -common <a,b,...> [options] (<file>|< <file>)

Extracts the "common (synchronized) events" of a given participant set from a sequence
diagram - the messages whose sender and receiver are both in that set - and prints them as a
TSV of (sender<TAB>receiver<TAB>label).
The output is deduplicated as an event set (in order of first appearance). A singleton set
(e.g. -common Foo) yields Foo's self-transitions.
Use it to pull out the common events of a participant set you plan to fold with pumlseqfold,
and pipe the result into pumlseqhide --msg-file.
For PNG input, the embedded PlantUML metadata is decoded.

Options:
  -common string
    	participant set to extract common events for (comma separated, at least one)
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqmsg -common B,C ./seq.puml
  B	C	fwd
  C	B	resp
```

### pumlseqproj

```console
$ pumlseqproj -h
Usage: pumlseqproj -ids <a,b,...> [options] (<file>|< <file>)

Keeps only the messages of a sequence diagram whose sender or receiver is in the given
participant set, removing every other message line (this is the projection tr ↾ A of the
trace, the dual of pumlseqhide).
Notes attached to removed messages are removed along with them, and control blocks left empty
are pruned by the same rule as pumlseqhide. After projecting, any participant/actor
declaration that is no longer referenced is dropped (declarations named in -ids, and those
still referenced by a remaining message, note, ref over, or activate/deactivate, are kept).
A box that no longer encloses any declaration is also removed. ref over is never removed and
is left untouched (to project through its referenced content too, first run pumlseqexpand).
It is an error if -ids names a participant that is absent from the diagram (neither declared
nor referenced).
Pass -name to set the name on the output's @startuml (omit it to keep the input's name).
For PNG input, the embedded PlantUML metadata is decoded.

Options:
  -ids string
    	participant ids to keep (comma separated, at least one)
  -name string
    	name to set on the output's @startuml (omit to keep the input's name)
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqproj -ids B ./seq.puml
  $ pumlseqproj -ids B,C ./seq.puml
  $ pumlseqproj -ids B -name Focused <./seq.puml
```

### pumlseqrefinement

```console
$ pumlseqrefinement -h
Usage: pumlseqrefinement [options] <abstract> <detail>

Checks whether detail is a refinement of abstract (CSP trace refinement:
traces(abs_φ(detail)) ⊆ traces(abstract)). Use -map to give the folding map φ
that collapses detail's participants onto abstract's participants (any participant
not listed maps to itself by identity). abs_φ(detail) is detail folded through φ;
messages internal to a folded box remain visible as self-transitions (hide them
first with pumlseqhide if that's what you want).
Either abstract or detail may be read from stdin using -.

Semantics: this is a simplified, message-level model inspired by UML and CSP, not
the full interaction semantics of either. A whole message is one atomic event (UML
models its send and receive as two separate occurrences; this tool does not).
Juxtaposition is weak sequencing in the UML sense of ordering per lifeline: an
event on a lifeline cannot occur until everything written before it on that same
lifeline has finished, events on lifelines that share no participant have no
ordering between them, and wrapping them in a block does not change that. alt/opt
is choice; loop is zero or more repetitions; par is full interleaving;
group/critical/break are transparent (critical and break carry meanings in UML that
this model does not reproduce). Note that pumlseqsubset uses a different, flatter
model in which alt branches are concatenated rather than treated as a choice.

Note: the more messages that share no participant, the more the state count grows
exponentially (about 1,000 states for 10 pairwise-disjoint messages, about 260,000
for 18). In a diagram where pumlseqbreak has confirmed there is no break in
causality, consecutive messages always share a participant, so this matches a
total order and no blow-up occurs.

A "ref over" spans N lifelines to invoke another diagram and cannot be represented
as a single message, so an unexpanded ref is never silently ignored — it aborts.
Run the diagram through pumlseqexpand first.

Exit codes: 0 = refinement holds / 1 = refinement fails (the shortest
counterexample trace is printed to stderr) / 2 = usage or input error.

Options:
  -map string
    	TSV folding map (detail ID<TAB>abstract ID); use - to read from stdin
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqrefinement -map map.tsv abstract.puml detail.puml
  $ pumlseqhide --msg-file - detail.puml <internal.tsv | pumlseqrefinement abstract.puml -
```

### pumlseqsubset

```console
$ pumlseqsubset -h
Usage: pumlseqsubset [options] <super> <sub>

Treats each sequence diagram as a sequence of (sender, receiver) pairs and
checks the relationship between <super> and <sub>. Message text and arrow
style are ignored. Exits 0 if the relationship holds, 1 if it doesn't, and
on failure lists the unmatched messages as TSV; the columns depend on -mode.
-mode exact (default) requires both diagrams to represent the same message
sequence, and reports the leftovers from either side as 3 columns:
(sub|super<TAB>sender<TAB>receiver). The first column says which diagram the
message appeared only in; sub's leftovers are listed in appearance order,
followed by super's.
-mode allow-subseq checks whether <sub> is a subsequence of <super> (i.e.
whether <super> can be obtained by inserting elements into <sub>), and
allows extra messages on the <super> side. (This mode is meant for cases
where <super> is a completed diagram and <sub> is the original, incomplete
one.) Any leftovers are necessarily on the <sub> side, so they're listed in
appearance order as 2 columns: (sender<TAB>receiver).
Either argument may be - to read from stdin.
Model: this is a structural comparison over one flat message sequence, not
an evaluation of UML's interaction semantics. Only par is given a meaning:
its branches run concurrently, so any linearization of them on either side
is allowed. Every other block (alt, opt, loop, group, critical, break) is
flattened by concatenating its branches in written order -- an alt is not
treated as a choice between branches, and a loop body is taken exactly
once. If you need alt as a choice and loop as repetition, use
pumlseqrefinement, which evaluates them as a CSP process instead.
Note that the number of states to check grows exponentially with the number
of par branches, multiplying several-fold per additional branch. Diagrams
with up to about 8 branches finish instantly, but a single par with 10 or
more branches can make the runtime impractical.
ref over blocks can't be represented as a message sequence, so an
unexpanded ref causes the command to exit with status 2 rather than being
silently ignored. Run it through pumlseqexpand first.
For PNG input, the embedded PlantUML metadata is read. Errors exit with
status 2.

Options:
  -mode string
    	comparison mode exact|allow-subseq (default exact) (default "exact")
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqsubset super.puml sub.puml; echo $?
  sub	B	X
  super	B	C
  super	C	B
  1

  $ pumlseqsubset -mode allow-subseq super.puml sub.puml; echo $?
  B	X
  1
```

### pumlseqtrace

```console
$ pumlseqtrace -h
Usage: pumlseqtrace -o <out_dir> [options] (<file>|< <file>)

Enumerates a sequence diagram's event traces and writes each one as a
single-column TSV file. An event is a single message, in the order it
appears in the diagram. Each TSV has one column, the message's label, and
one event per line.
File names follow <out_dir>/<basename>_trace_<idx>.tsv, where idx starts at
1 and is zero-padded to the number of digits in the trace count. Written
paths are printed to stdout, one per line.

Per-block interpretation:
  alt      one trace per compartment.
  opt      two traces: one that runs the body, one that doesn't.
  loop     one trace for each of 1..<loop-unroll> repetitions
           (-loop-unroll 0 unrolls to 0 repetitions only).
  par      enumerates the full interleaving of its operands, keeping the
           order within each operand.
  break    runs up to the body and then cuts the trace short (nothing
           after it runs). A break inside a par cuts the whole trace short
           at the point that event occurs, so events still pending in
           other, concurrent operands never run either.
  group and other blocks with no control structure are transparent.

par is prone to a trace-count explosion. When the count is expected to
exceed 100, a warning is printed to stderr; interrupt with Ctrl+C and add
-interleave-sampling if you want to cut it down (enumeration continues
after the warning either way).
Pass -interleave-sampling <rate> to sample traces down, keeping each one
with probability rate.
The seed is fixed, so the same input and arguments always produce the same
result (override it with -seed).

basename defaults to the input file name with a trailing .png or .puml
stripped. When reading from stdin, -basename is required, and a -basename
containing a path separator (/ \) is an error.
<out_dir> is created if it does not exist, and files with the same name
are overwritten. Every file is written to a temporary file first and only
then renamed into place, so a failure while generating or writing content
leaves <out_dir> untouched.
An unexpanded ref over is an error (a call to another diagram can't be
represented as a single event, so run pumlseqexpand first).
For PNG input, the embedded PlantUML metadata is read.

Options:
  -basename string
    	base name for output files (default: derived from the input file name; required when reading from stdin)
  -interleave-sampling float
    	probability of keeping each trace (0 < rate <= 1; unset means no sampling)
  -loop-unroll int
    	maximum number of loop unrollings (enumerates each of 1..N repetitions; 0 skips the body entirely) (default 1)
  -o string
    	directory to write the trace TSV files into (created if it does not exist)
  -seed uint
    	seed for the pseudo-random generator used for sampling (default 1)
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumlseqtrace -o ./out ./seq.puml
  ./out/seq_trace_1.tsv
  ./out/seq_trace_2.tsv

  $ pumlseqtrace -o ./out -basename flow -interleave-sampling 0.01 <./seq.puml
  ./out/flow_trace_1.tsv
```

### pumltodo

```console
$ pumltodo -h
Usage: pumltodo [options] (<file>|< <file>)

Extracts TODOs matching --regexp from PlantUML line comments ('), block comments
(/' '/), note/hnote/rnote, and ref blocks, and lists them as 3-column TSV: file
name, line number, content.
For a multi-line block, a single matching line counts the whole block as one hit.
If the input is a PNG, the PlantUML embedded in its metadata is parsed instead.

Options:
  -regexp string
    	regular expression to match (default "(TODO|todo)")
  -v	print version and exit
  -version
    	print version and exit

Examples:
  $ pumltodo ./diagram.puml
  ./diagram.puml	3	TODO: ...

  $ pumltodo --regexp '(TODO|FIXME)' <./diagram.puml
  -	3	TODO: ...
```

## References

These tools borrow vocabulary from the specifications below. Where a tool
simplifies or departs from one, its `-h` output says so — in particular,
`pumlseqrefinement` and `pumlseqsubset` use message-level models that are
inspired by UML and CSP rather than implementations of their full semantics,
and the two use *different* models from each other.

- [PlantUML](https://plantuml.com/) — the diagram syntax these tools read and write; see the [FAQ](https://plantuml.com/faq) for the PlantUML source embedded in PNG metadata.
- [C4 model](https://c4model.com/) — the level vocabulary used by `pumlseq2c4`, `pumlc4fold`, `pumlc4bdyls` and `pumlc42seqchk`, in particular the [System Context diagram](https://c4model.com/diagrams/system-context) (Lv.1) and the [Container diagram](https://c4model.com/diagrams/container) (Lv.2).
- [C4-PlantUML](https://github.com/plantuml-stdlib/C4-PlantUML) — the `C4_Context` / `C4_Container` / `C4_Component` macro set the C4 output includes.
- [OMG UML 2.5.1](https://www.omg.org/spec/UML/2.5.1/About-UML) — the source of the sequence diagram terms (lifeline, message, weak sequencing, `alt` / `opt` / `loop` / `par` / `critical` / `break`).
- [Hoare, *Communicating Sequential Processes*](https://www.cs.ox.ac.uk/bill.roscoe/publications/68b.pdf) and the [FDR manual](https://cocotec.io/fdr/manual/) — traces refinement, hiding and trace restriction, as used by `pumlseqrefinement`, `pumlseqhide` and `pumlseqproj`.
