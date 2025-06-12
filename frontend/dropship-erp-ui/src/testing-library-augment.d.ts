import '@testing-library/react';
import {fireEvent, screen, waitFor, within} from '@testing-library/dom';

declare module '@testing-library/react' {
  export {fireEvent, screen, waitFor, within};
}
