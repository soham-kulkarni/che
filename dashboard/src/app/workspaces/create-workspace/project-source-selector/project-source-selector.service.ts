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
import {ProjectSource} from './project-source.enum';
import {TemplateSelectorSvc} from './template-selector/template-selector.service';
import {ImportBlankProjectService} from './import-blank-project/import-blank-project.service';
import {ImportGitProjectService} from './import-git-project/import-git-project.service';
import {ImportZipProjectService} from './import-zip-project/import-zip-project.service';
import {Observable} from '../../../../components/utils/observable';

/**
 * This class is handling the service for the project selector.
 *
 * @author Oleksii Kurinnyi
 */
export class ProjectSourceSelectorService extends Observable {
  /**
   * Template selector service.
   */
  private templateSelectorSvc: TemplateSelectorSvc;
  /**
   * Import blank project service.
   */
  private importBlankProjectService: ImportBlankProjectService;
  /**
   * Import Git project service.
   */
  private importGitProjectService: ImportGitProjectService;
  /**
   * Import Zip project service.
   */
  private importZipProjectService: ImportZipProjectService;
  /**
   * Project templates to import.
   */
  private projectTemplates: Array<che.IProjectTemplate>;

  /**
   * Default constructor that is using resource injection
   * @ngInject for Dependency injection
   */
  constructor(templateSelectorSvc: TemplateSelectorSvc, importBlankProjectService: ImportBlankProjectService, importGitProjectService: ImportGitProjectService, importZipProjectService: ImportZipProjectService) {
    super();

    this.templateSelectorSvc = templateSelectorSvc;
    this.importBlankProjectService = importBlankProjectService;
    this.importGitProjectService = importGitProjectService;
    this.importZipProjectService = importZipProjectService;

    this.projectTemplates = [];
  }

  /**
   * Builds new project template based on blank project template.
   *
   * @param {any} props
   * @return {che.IProjectTemplate}
   */
  buildProjectTemplate(props: any): che.IProjectTemplate {
    const blankProjectTemplate = this.templateSelectorSvc.getTemplateByName('blank-project');
    return angular.merge({}, blankProjectTemplate, props);
  }

  /**
   * Adds project from source.
   *
   * @param {ProjectSource} source the project's source
   */
  addProjectTemplateFromSource(source: ProjectSource): void {
    // move to addProjectTemplateAs
    switch (source) {
      case ProjectSource.SAMPLES:
        const projectTemplates = this.templateSelectorSvc.getTemplates();
        projectTemplates.forEach((projectTemplate: che.IProjectTemplate) => {
          this.addProjectTemplate(projectTemplate);
        });
        break;
      case ProjectSource.BLANK: {
        const projectProps = this.importBlankProjectService.getProjectProps();
        const projectTemplate = this.buildProjectTemplate(projectProps);
        this.addProjectTemplate(projectTemplate);
      }
        break;
      case ProjectSource.GIT: {
        const projectProps = this.importGitProjectService.getProjectProps();
        const projectTemplate = this.buildProjectTemplate(projectProps);
        this.addProjectTemplate(projectTemplate);
      }
        break;
      case ProjectSource.GITHUB:
        break;
      case ProjectSource.ZIP: {
        const projectProps = this.importZipProjectService.getProjectProps();
        const projectTemplate = this.buildProjectTemplate(projectProps);
        this.addProjectTemplate(projectTemplate);
      }
        break;
    }
  }

  /**
   * Adds project template to the list.
   *
   * @param {che.IProjectTemplate} projectTemplate the project template
   */
  addProjectTemplate(projectTemplate: che.IProjectTemplate): void {
    if (this.isProjectTemplateNameUnique(projectTemplate.name) === false) {
      return;
    }

    this.projectTemplates.push(projectTemplate);
  }

  /**
   * Returns list of project templates.
   *
   * @return {Array<che.IProjectTemplate>}
   */
  getProjectTemplates(): Array<che.IProjectTemplate> {
    return this.projectTemplates;
  }

  /**
   * Returns <code>true</code> if project's template name is unique.
   *
   * @param {string} name the project's template name
   * @return {boolean}
   */
  isProjectTemplateNameUnique(name: string): boolean {
    return this.projectTemplates.every((projectTemplate: che.IProjectTemplate) => {
      return projectTemplate.name !== name;
    });
  }

}
