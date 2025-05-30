YAML Reference
==============

iapetus workflows can be defined in YAML for easy, no-code automation.

YAML Schema
-----------
.. code-block:: yaml

   name: my-wf
   backend: bash
   env_map:
     FOO: bar
   steps:
     - name: hello
       command: echo
       args: ["hello"]
       timeout: 5s
       backend: docker
       image: alpine:3.18
       env_map:
         BAR: baz
       raw_asserts:
         - output_contains: hello

Supported assertion types:
-------------------------
- exit_code: 0
- output_equals: "foo"
- output_contains: "bar"
- output_json_equals: '{"foo": 1}'
- output_matches_regexp: '^foo.*$'
- skip_json_nodes: ["foo.bar"] (for JSON assertions)

Tips:
-----
- Indent with spaces, not tabs
- Use quotes for strings with special characters
- See :doc:`tutorials` for more examples

See also: :doc:`api` 