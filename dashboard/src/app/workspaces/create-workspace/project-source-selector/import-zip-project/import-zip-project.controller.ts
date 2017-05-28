/*
 * Copyright (c) 2015-2017 Codenvy, S.A.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *   Codenvy, S.A. - initial API and implementation
 */
'use strict';

import {ImportZipProjectService} from './import-zip-project.service';

/**
 * This class is handling the controller for the Zip project import.
 *
 * @author Oleksii Kurinnyi
 */
export class ImportZipProjectController {
  /**
   * Import Zip project service.
   */
  private importZipProjectService: ImportZipProjectService;
  /**
   * Git repository location.
   */
  private location: string;

  /**
   * Default constructor that is using resource injection
   * @ngInject for Dependency injection
   */
  constructor(importZipProjectService: ImportZipProjectService) {
    this.importZipProjectService = importZipProjectService;
  }

  /**
   * Callback which is called when location is changed.
   */
  onLocationChanged(): void {
    this.importZipProjectService.onLocationChanged(this.location);
  }
}
