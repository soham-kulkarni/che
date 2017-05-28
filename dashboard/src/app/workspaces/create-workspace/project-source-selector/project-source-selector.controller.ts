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
import {ProjectSourceSelectorService} from './project-source-selector.service';
import {ProjectSource} from './project-source.enum';

/**
 * This class is handling the controller for the project selector.
 *
 * @author Oleksii Kurinnyi
 */
export class ProjectSourceSelectorController {
  /**
   * Project selector service.
   */
  private projectSourceSelectorService: ProjectSourceSelectorService;
  /**
   * Available project sources.
   */
  private projectSource: Object;
  /**
   * Selected project's source.
   */
  private selectedSource: ProjectSource;

  /**
   * Default constructor that is using resource injection
   * @ngInject for Dependency injection
   */
  constructor($timeout, projectSourceSelectorService: ProjectSourceSelectorService) {
    this.$timeout = $timeout;
    this.projectSourceSelectorService = projectSourceSelectorService;

    this.projectSource = ProjectSource;
    this.selectedSource = ProjectSource.SAMPLES;
  }

  /**
   * Returns list of project templates which are ready to be imported.
   *
   * @return {Array<che.IProjectTemplate>}
   */
  getProjectTemplates(): Array<che.IProjectTemplate> {
    return this.projectSourceSelectorService.getProjectTemplates();
  }

  /**
   * Add project template from selected source to the list.
   */
  addProjectTemplate(): void {
    this.projectSourceSelectorService.addProjectTemplateFromSource(this.selectedSource);
  }

  // todo remove this method
  test(state) {
    console.log('>>> ProjectSourceSelectorController.test');
    console.log('>>> state: ', state);
    this.$timeout(() => {
      this.isOpen = state;
      console.log('>>> this.isOpen: ', this.isOpen);
    }, 50)
  }
}
