import { CdkCellDef } from '@angular/cdk/table';
import { Directive, Input } from '@angular/core';
import { MatCellDef, MatTableDataSource } from '@angular/material/table';
import { Observable } from 'rxjs';

@Directive({
  selector: '[matCellDef]', // same selector as MatCellDef
  providers: [{ provide: CdkCellDef, useExisting: TypeSafeMatCellDef }],
})
export class TypeSafeMatCellDef<T> extends MatCellDef {
  @Input() matCellDefDataSource?: T[] | Observable<T[]> | MatTableDataSource<T>;

  static ngTemplateContextGuard<T>(
    dir: TypeSafeMatCellDef<T>,
    ctx: any,
  ): ctx is { $implicit: T; index: number } {
    return true;
  }
}
