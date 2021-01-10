import { createLogger, LogLevelString } from 'bunyan';
import * as database from '../clients/database';
import { BatchProcessor } from './batch-processor';
import { IDBBatch } from './interfaces';
import * as utils from './utils';

const log = createLogger({ name: 'lib/batch-manager.ts', level: utils.constants.LOG_LEVEL as LogLevelString });

/**
 * Lifecycle manager for BatchProcessor instances, within a single type, across multiple authors
 */
export class BatchManager<IRecordType> {

  processors: {[author: string]: BatchProcessor<IRecordType>} = {};

  constructor(
    private type: string,
    private processBatchCallback: (batch: IDBBatch<IRecordType>) => Promise<void>,
  ) { }

  public async init() {
    // Query all incomplete batches for our type, in creation order
    const inflightBatches = await database.retrieveBatches({
      type: this.type,
      completed: null,
    }, 0, 0, { created: 1 });
    const byAuthor: {[author: string]: IDBBatch<IRecordType>[]} = {};
    for (const inflight of inflightBatches) {
      const forAuthor = byAuthor[inflight.author] = byAuthor[inflight.author] || [];
      forAuthor.push(inflight);
    }
    // Init a set of processors for each distinct authors.
    // Note these will be reaped if no work comes in while we're processing the backlog
    for (const [author, forAuthor] of Object.entries(byAuthor)) {
      await this.getProcessor(author).init(forAuthor);
    }
  }

  protected processorCompleteCallback(author: string) {
    log.trace(`${this.type} batch manager: Reaping processor for ${author}`);
    delete this.processors[author];
  }

  public getProcessor(author: string) {
    if (!this.processors[author]) {
      log.trace(`${this.type} batch manager: Creating processor for ${author}`);
      this.processors[author] = new BatchProcessor<IRecordType>(
        author,
        this.type,
        this.processBatchCallback,
        this.processorCompleteCallback.bind(this),
      );
    }
    return this.processors[author];
  }

}