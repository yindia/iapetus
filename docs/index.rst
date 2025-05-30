.. iapetus documentation master file

Welcome to iapetus's documentation!
===================================

**iapetus** is an open source, extensible workflow orchestrator for automating and testing command-line tasks, DevOps pipelines, and CI/CD workflows. It supports parallel DAG execution, plugin backends (bash, docker, and more), YAML and Go configuration, and robust assertions.

.. note::
   This project is under active development. Feedback and contributions are welcome!

Project Links
-------------

- `GitHub Repository <https://github.com/yindia/iapetus>`_
- `Main README <../README.md>`_
- `Usage Guide <../USAGE.md>`_
- `API & YAML Reference <../REFERENCE.md>`_
- `Contributing Guide <../CONTRIBUTING.md>`_

Installation
------------

.. code-block:: shell

   git clone https://github.com/yindia/iapetus.git
   cd iapetus
   go mod tidy

Contents
--------

.. toctree::
   :maxdepth: 2
   :caption: Documentation

   Home <self>
   usage
   api
   contributing

Getting Started
---------------

See the :doc:`usage` section for a quickstart, workflow/task examples, and best practices.

For advanced usage, see the :doc:`api` reference.

If you'd like to contribute, see :doc:`contributing`.