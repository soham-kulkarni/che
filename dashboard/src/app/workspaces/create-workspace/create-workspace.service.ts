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

import {CheWorkspace} from '../../../components/api/che-workspace.factory';
import IdeSvc from '../../ide/ide.service';
import {NamespaceSelectorSvc} from './namespace-selector/namespace-selector.service';
import {StackSelectorSvc} from './stack-selector/stack-selector.service';
import {TemplateSelectorSvc} from './project-selector/template-selector/template-selector.service';

/**
 * todo
 *
 * @author Oleksii Kurinnyi
 */
export class CreateWorkspaceSvc {
  /**
   * Location service.
   */
  private $location: ng.ILocationService;
  /**
   * Promises service.
   */
  private $q: ng.IQService;
  /**
   * Workspace API interaction.
   */
  private cheWorkspace: CheWorkspace;
  /**
   * IDE service.
   */
  private ideSvc: IdeSvc;
  /**
   * Namespace selector service.
   */
  private namespaceSelectorSvc: NamespaceSelectorSvc;
  /**
   * Stack selector service.
   */
  private stackSelectorSvc: StackSelectorSvc;
  /**
   * Template selector service.
   */
  private templateSelectorSvc: TemplateSelectorSvc;
  /**
   * The list of workspaces by namespace.
   */
  private workspacesByNamespace: {
    [namespaceId: string]: Array<che.IWorkspace>
  };

  // todo
  private workspaceOfProject: any;
  // todo
  private namespace: string;
  // todo
  private project: any;
  // todo
  private ideAction: any;

  /**
   * Default constructor that is using resource injection
   * @ngInject for Dependency injection
   */
  constructor($location: ng.ILocationService, $q: ng.IQService, cheWorkspace: CheWorkspace, ideSvc: IdeSvc, namespaceSelectorSvc: NamespaceSelectorSvc, stackSelectorSvc: StackSelectorSvc, templateSelectorSvc: TemplateSelectorSvc) {
    this.$location = $location;
    this.$q = $q;
    this.cheWorkspace = cheWorkspace;
    this.ideSvc = ideSvc;
    this.namespaceSelectorSvc = namespaceSelectorSvc;
    this.stackSelectorSvc = stackSelectorSvc;
    this.templateSelectorSvc = templateSelectorSvc;

    this.workspacesByNamespace = {};
  }

  /**
   * Fills in list of workspace's name in current namespace,
   * and triggers validation of entered workspace's name
   *
   * @param {string} namespaceId a namespace ID
   * @return {IPromise<any>}
   */
  fetchWorkspacesByNamespace(namespaceId: string): ng.IPromise<any> {
    return this.getOrFetchWorkspacesByNamespace(namespaceId).then((workspaces: Array<che.IWorkspace>) => {
      return this.$q.when(workspaces);
    }, (error: any) => {
      // user is not authorized to get workspaces by namespace
      return this.getOrFetchWorkspaces();
    }).then((workspaces: Array<che.IWorkspace>) => {
      this.workspacesByNamespace[namespaceId] = workspaces;
      return this.$q.when(workspaces);
    });
  }

  /**
   * Returns promise for getting list of workspaces by namespace.
   *
   * @param {string} namespaceId a namespace ID
   * @return {ng.IPromise<any>}
   */
  getOrFetchWorkspacesByNamespace(namespaceId: string): ng.IPromise<any> {
    const defer = this.$q.defer();

    const workspacesByNamespaceList = this.cheWorkspace.getWorkspacesByNamespace(namespaceId) || [];
    if (workspacesByNamespaceList.length) {
      defer.resolve(workspacesByNamespaceList);
    } else {
      this.cheWorkspace.fetchWorkspacesByNamespace(namespaceId).then(() => {
        defer.resolve(this.cheWorkspace.getWorkspacesByNamespace(namespaceId) || []);
      }, (error: any) => {
        // not authorized
        defer.reject(error);
      });
    }

    return defer.promise;
  }

  /**
   * Returns promise for getting list of workspaces owned by user
   *
   * @return {ng.IPromise<any>}
   */
  getOrFetchWorkspaces(): ng.IPromise<any> {
    const defer = this.$q.defer();
    const workspacesList = this.cheWorkspace.getWorkspaces();
    if (workspacesList.length) {
      defer.resolve(workspacesList);
    } else {
      this.cheWorkspace.fetchWorkspaces().finally(() => {
        defer.resolve(this.cheWorkspace.getWorkspaces());
      });
    }

    return defer.promise;
  }

  createWorkspace(workspaceConfig: che.IWorkspaceConfig): ng.IPromise<any> {
    console.log('>>> stackId: ', this.stackSelectorSvc.getStackId());
    console.log('>>> workspaceConfig: ', workspaceConfig);
    console.log('>>> templateNames: ', this.templateSelectorSvc.getTemplateNames());

    const namespaceId = this.namespaceSelectorSvc.getNamespaceId(),
          templateNames = this.templateSelectorSvc.getTemplateNames();

    return this.cheWorkspace.createWorkspaceFromConfig(namespaceId, workspaceConfig, {}).then((workspace: che.IWorkspace) => {
      console.log('>>> new workspace created');

      this.cheWorkspace.startWorkspace(workspace.id, workspace.config.defaultEnv).then(() => {
        this.redirectToIde(namespaceId, workspace);
        console.log('>>> ws is starting');
        this.cheWorkspace.getWorkspacesById().set(workspace.id, workspace);
        this.cheWorkspace.startUpdateWorkspaceStatus(workspace.id);
        // this.ideSvc.openIde(workspace.id);
        return this.cheWorkspace.fetchStatusChange(workspace.id, 'RUNNING');
      }).then(() => {
        console.log('>>> workspace is running');
        return this.cheWorkspace.fetchWorkspaceDetails(workspace.id);
      }).then(() => {
        console.log('>>> workspace is fetched');
        return this.addProjects(workspace.id, templateNames);
      });
    });
  }

  redirectToIde(namespaceId: string, workspace: che.IWorkspace): void {
    const path = `/ide/${namespaceId}/${workspace.config.name}`;
    this.$location.path(path);
  }

  addProjects(workspaceId: string, templateNames: string[]): ng.IPromise<any> {
    console.log('>>> CreateWorkspaceSvc.addProjects, arguments: ', arguments);
    if (templateNames.length === 0) {
      return;
    }

    const projects = this.templateSelectorSvc.getTemplates().filter((template: che.IProjectTemplate) => {
      return templateNames.indexOf(template.name) !== -1;
    });
    console.log('>>> projects: ', projects);

    const workspaceAgent = this.cheWorkspace.getWorkspaceAgent(workspaceId);
    console.log('>>> workspaceAgent: ', workspaceAgent);
    return workspaceAgent.getProject().createProjects(projects).then(() => {
      console.log('>>> before importing projects');
      return this.importProjects(workspaceId, projects);
    });
  }

  importProjects(workspaceId: string, projects: Array<che.IProjectTemplate>): ng.IPromise<any> {
    console.log('>>> CreateWorkspaceSvc.importProjects, arguments: ', arguments);
    const defer = this.$q.defer();
    defer.resolve();
    let accumulatorPromise = defer.promise;

    const projectTypeResolverService = this.cheWorkspace.getWorkspaceAgent(workspaceId).getProjectTypeResolver();

    accumulatorPromise = projects.reduce((_accumulatorPromise: ng.IPromise<any>, project: che.IProjectTemplate, index: number) => {
      return _accumulatorPromise.then(() => {
        return this.addCommands(workspaceId, project.name, project.commands).finally(() => {
          console.log('>>> commands are added for project.name: ', project.name);
          return projectTypeResolverService.resolveProjectType(project as any).finally(() => {
            console.log('>>> import is done for project.name: ', project.name);
          });
        });
      });
    }, accumulatorPromise);

    return accumulatorPromise;
  }

  addCommands(workspaceId: string, projectName: string, projectCommands: any[]): ng.IPromise<any> {
    console.log('>>> CreateWorkspaceSvc.addCommands');
    const defer = this.$q.defer();
    defer.resolve();
    let accumulatorPromise = defer.promise;

    accumulatorPromise = projectCommands.reduce((_accumulatorPromise: ng.IPromise<any>, command: any, number: number) => {
      console.log('>>> add command number: ', number);
      command.name = projectName + ':' + command.name;
      return _accumulatorPromise.then(() => {
        return this.cheWorkspace.addCommand(workspaceId, command).finally(() => {
          console.log(`>>> added command ${command.name} in project ${projectName}`);
        });
      });
    }, accumulatorPromise);

    return accumulatorPromise;
  }

  // review methods below
  // whether they are necessary or not

  setWorkspaceOfProject(workspaceOfProject: any): void {
    this.workspaceOfProject = workspaceOfProject;
  }

  getWorkspaceOfProject(): void {
    return this.workspaceOfProject;
  }

  setWorkspaceNamespace(namespace: string): void {
    this.namespace = namespace;
  }

  getWorkspaceNamespace(): string {
    return this.namespace;
  }

  setProject(project: any): void {
    this.project = project;
  }

  getProject(): any {
    return this.project;
  }

  hasIdeAction(): boolean {
    return this.getIDEAction().length > 0;
  }

  getIDEAction(): any {
    return this.ideAction;
  }

  setIDEAction(ideAction: any): void {
    this.ideAction = ideAction;
  }

  getIDELink(): string {
    let link = '#/ide/' + this.getWorkspaceNamespace() + '/' + this.getWorkspaceOfProject();
    if (this.hasIdeAction()) {
      link = link + '?action=' + this.ideAction;
    }
    return link;
  }

  _redirectToIDE(): void {
    const path = '/ide/' + this.getWorkspaceNamespace() + '/' + this.getWorkspaceOfProject();
    this.$location.path(path);

    if (this.getIDEAction()) {
      this.$location.search({'action': this.getIDEAction()});
    }
  }

}
