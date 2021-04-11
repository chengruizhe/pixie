<?php

/**
 * This file is automatically generated. Use 'arc liberate' to rebuild it.
 *
 * @generated
 * @phutil-library-version 2
 */
phutil_register_library_map(array(
  '__library_version__' => 2,
  'class' => array(
    'ArcanistGoImportsLinter' => 'lint/ArcanistGoImportsLinter.php',
    'ArcanistGolangCiLinter' => 'lint/ArcanistGolangCiLinter.php',
    'ExpCheckerTestEngine' => 'unit/ExpCheckerTestEngine.php',
    'FileCheckerTestEngine' => 'unit/FileCheckerTestEngine.php',
    'GazelleCheckerTestEngine' => 'unit/GazelleCheckerTestEngine.php',
    'GoGenerateCheckerTestEngine' => 'unit/GoGenerateCheckerTestEngine.php',
    'PLTestEngine' => 'unit/PLTestEngine.php',
    'PrototoolCheckerTestEngine' => 'unit/PrototoolCheckerTestEngine.php',
  ),
  'function' => array(),
  'xmap' => array(
    'ArcanistGoImportsLinter' => 'ArcanistExternalLinter',
    'ArcanistGolangCiLinter' => 'ArcanistExternalLinter',
    'PLTestEngine' => 'ArcanistUnitTestEngine',
  ),
));
