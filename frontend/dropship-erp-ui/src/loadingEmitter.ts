export class LoadingEmitter extends EventTarget {
  private _count = 0;

  start() {
    this._count++;
    this.dispatchEvent(new CustomEvent('change', { detail: this._count }));
  }

  end() {
    this._count = Math.max(0, this._count - 1);
    this.dispatchEvent(new CustomEvent('change', { detail: this._count }));
  }
}

export const loadingEmitter = new LoadingEmitter();
